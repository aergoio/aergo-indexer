package esindexer

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/olivere/elastic"
	"golang.org/x/sync/errgroup"
)

// BulkIndexer is a utility function that uses a generator function to create ES documents and inserts them in chunks
func BulkIndexer(ctx context.Context, logger *log.Logger, client *elastic.Client, channel chan EsType, generator func() error, indexName string, typeName string, chunkSize int) {
	// Setup a group of goroutines
	// The first goroutine will emit documents and send it to the second goroutine via the channel.
	// The second goroutine will simply bulk insert the documents.
	g, ctx := errgroup.WithContext(ctx)

	begin := time.Now()

	// Firs goroutine to create documents by adding to channel
	g.Go(generator)

	// Second goroutine consumes the documents sent from the first and bulk insert into ES
	var total uint64
	g.Go(func() error {
		bulk := client.Bulk().Index(indexName).Type(typeName)
		for d := range channel {
			atomic.AddUint64(&total, 1)
			bulk.Add(elastic.NewBulkIndexRequest().Id(d.GetID()).Doc(d))
			if bulk.NumberOfActions() >= chunkSize {
				res, err := bulk.Do(ctx)
				if err != nil {
					return err
				}
				if res.Errors {
					return errors.New("bulk commit failed")
				}
			}

			select {
			default:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Commit the final batch before exiting
		if bulk.NumberOfActions() > 0 {
			_, err := bulk.Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})

	// Wait until all goroutines are finished
	if err := g.Wait(); err != nil {
		logger.Warn().Err(err).Msg(fmt.Sprintf("Error bulk indexing %ss", typeName))
	}

	// Final results
	dur := time.Since(begin).Seconds()
	pps := int64(float64(total) / dur)
	logger.Info().Uint64("total", total).Int64("perSecond", pps).Msg(fmt.Sprintf("Done bulk indexing %ss", typeName))
}
