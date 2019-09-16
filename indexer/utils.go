package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/aergoio/aergo-indexer/indexer/db"
	doc "github.com/aergoio/aergo-indexer/indexer/documents"

	"github.com/aergoio/aergo-lib/log"
	"golang.org/x/sync/errgroup"
)

// BulkIndexer is a utility function that uses a generator function to create ES documents and inserts them in chunks
func BulkIndexer(ctx context.Context, logger *log.Logger, dbController db.DbController, channel chan doc.DocType, generator func() error, indexName string, typeName string, chunkSize int, upsert bool) {
	// Setup a group of goroutines
	g, ctx := errgroup.WithContext(ctx)

	begin := time.Now()
	var total uint64

	// First goroutine to create documents by adding to channel
	g.Go(generator)

	// Second goroutine consumes the documents sent from the first and bulk insert into ES
	g.Go(func() error {
		_total, err := dbController.InsertBulk(channel, db.UpdateParams{IndexName: indexName, TypeName: typeName, Size: chunkSize, Upsert: upsert})
		if err != nil {
			return err
		}
		total = _total
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
