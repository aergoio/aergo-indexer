package indexer

import (
	"io"

	"github.com/aergoio/aergo-indexer/indexer/db"
	doc "github.com/aergoio/aergo-indexer/indexer/documents"
)

type esBlockNo struct {
	*doc.BaseEsType
	BlockNo uint64 `json:"no" db:"no"`
}

// CheckConsistency gets all block numbers from 0 to ns.lastBlockHeight in order and checks for "holes"
func (ns *Indexer) CheckConsistency() {
	count, err := ns.db.Count(db.QueryParams{IndexName: ns.indexNamePrefix + "block"})
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to query block count")
		return
	}
	if uint64(count) >= ns.lastBlockHeight+1 {
		ns.log.Info().Int64("total indexed", count).Uint64("expected", ns.lastBlockHeight+1).Msg("Skipping consistency check")
		return
	}
	ns.log.Info().Int64("total indexed", count).Uint64("expected", ns.lastBlockHeight+1).Msg("Checking consistency")

	prevBlockNo := uint64(0)
	missingBlocks := uint64(0)

	scroll := ns.db.Scroll(db.QueryParams{
		IndexName:    ns.indexNamePrefix + "block",
		TypeName:     "block",
		SelectFields: []string{"no"},
		Size:         10000,
		SortField:    "no",
		SortAsc:      true,
	}, func() doc.DocType {
		block := new(esBlockNo)
		return block
	})

	var checked int

	for {
		block, err := scroll.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			ns.log.Warn().Err(err).Msg("Failed to query block numbers")
			break
		}
		blockNo := block.(*esBlockNo).BlockNo
		if blockNo > prevBlockNo+1 {
			missingBlocks = missingBlocks + (blockNo - prevBlockNo - 1)
			ns.IndexBlocksInRange(prevBlockNo+1, blockNo-1)
		}
		prevBlockNo = blockNo
		checked++
	}

	ns.log.Info().Uint64("missing", missingBlocks).Int("checked", checked).Msg("Done with consistency check")
}
