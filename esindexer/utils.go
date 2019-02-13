package esindexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aergoio/aergo-lib/log"
	"github.com/olivere/elastic"
	"golang.org/x/sync/errgroup"
)

func logBulkResponse(logger *log.Logger, res *elastic.BulkResponse) {
	if res.Errors {
		for _, v := range res.Items {
			for action, item := range v {
				if item.Error != nil {
					resJSON, _ := json.Marshal(item.Error)
					logger.Error().Str("type", item.Type).Str("action", action).Str("id", item.Id).Msg(string(resJSON))
				}
			}
		}
	}
}

// BulkIndexer is a utility function that uses a generator function to create ES documents and inserts them in chunks
func BulkIndexer(ctx context.Context, logger *log.Logger, client *elastic.Client, channel chan EsType, generator func() error, indexName string, typeName string, chunkSize int, upsert bool) {
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
			if upsert {
				bulk.Add(elastic.NewBulkIndexRequest().Id(d.GetID()).Doc(d))
			} else {
				bulk.Add(elastic.NewBulkUpdateRequest().Id(d.GetID()).Doc(d).DocAsUpsert(true))
			}
			if bulk.NumberOfActions() >= chunkSize {
				dur := time.Since(begin).Seconds()
				pps := int64(float64(total) / dur)
				logger.Info().Int("chunkSize", chunkSize).Uint64("total", total).Int64("perSecond", pps).Msg(fmt.Sprintf("Commiting bulk chunk %ss", typeName))
				res, err := bulk.Do(ctx)
				logBulkResponse(logger, res)
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
			dur := time.Since(begin).Seconds()
			pps := int64(float64(total) / dur)
			logger.Info().Int("chunkSize", chunkSize).Uint64("total", total).Int64("perSecond", pps).Msg(fmt.Sprintf("Commiting bulk chunk %ss", typeName))
			res, err := bulk.Do(ctx)
			logBulkResponse(logger, res)
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
