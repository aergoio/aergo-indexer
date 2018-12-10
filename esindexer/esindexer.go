package esindexer

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aergoio/aergo-esindexer/types"
	"github.com/aergoio/aergo-lib/log"
	"github.com/mr-tron/base58/base58"
	"github.com/olivere/elastic"
)

type EsIndexer struct {
	client          *elastic.Client
	grpcClient      types.AergoRPCServiceClient
	aliasNamePrefix string
	indexNamePrefix string
	lastBlockHeight uint64
	lastBlockHash   string
	log             *log.Logger
	reindexing      bool
	esURL           string
}

// NewEsIndexer createws new EsIndexer instance
func NewEsIndexer(logger *log.Logger, esURL string, namePrefix string) *EsIndexer {
	aliasNamePrefix := namePrefix
	svc := &EsIndexer{
		esURL:           esURL,
		aliasNamePrefix: aliasNamePrefix,
		indexNamePrefix: generateIndexPrefix(aliasNamePrefix),
		lastBlockHeight: 0,
		lastBlockHash:   "",
		log:             logger,
		reindexing:      false,
	}
	return svc
}

func generateIndexPrefix(aliasNamePrefix string) string {
	return fmt.Sprintf("%s%s_", aliasNamePrefix, time.Now().UTC().Format("2006-01-02t15:04:05z"))
}

// CreateIndexIfNotExists creates the indices and aliases in ES
func (ns *EsIndexer) CreateIndexIfNotExists(documentType string) {
	initialized := true
	aliasName := ns.aliasNamePrefix + documentType
	ctx := context.Background()
	// Check for existing index to find out current indexNamePrefix
	if !ns.reindexing {
		res, err := ns.client.Aliases().Index("_all").Do(ctx)
		if err != nil {
			ns.log.Warn().Err(err).Msg("Error when checking for alias")
		}
		indices := res.IndicesByAlias(aliasName)
		if len(indices) > 0 {
			indexName := indices[0]
			ns.log.Info().Str("aliasName", aliasName).Str("indexName", indexName).Msg("Alias found")
			ns.indexNamePrefix = strings.TrimRight(indices[0], documentType)
		} else {
			initialized = false
			ns.reindexing = false
		}
	}
	// Create new index
	if ns.reindexing || !initialized {
		indexName := ns.indexNamePrefix + documentType
		ns.log.Info().Str("aliasName", aliasName).Str("indexName", indexName).Msg("Initializing alias")
		createIndex, err := ns.client.CreateIndex(indexName).BodyString(mappings[documentType]).Do(ctx)
		if err != nil || !createIndex.Acknowledged {
			ns.log.Warn().Err(err).Msg("Error when creating index")
		}
		ns.log.Info().Str("indexName", indexName).Msg("Created index")
		// Update alias, only when not reindexing
		if !ns.reindexing {
			err = ns.UpdateAlias(aliasName, indexName)
			if err != nil {
				ns.log.Warn().Err(err).Str("aliasName", aliasName).Str("indexName", indexName).Msg("Error when updating alias")
			} else {
				ns.log.Info().Str("aliasName", aliasName).Str("indexName", indexName).Msg("Updated alias")
			}
		}
	}
}

// UpdateAlias removes existing indexes and adds new index to alias
func (ns *EsIndexer) UpdateAlias(aliasName string, indexName string) error {
	ctx := context.Background()
	svc := ns.client.Alias()
	res, err := ns.client.Aliases().Index("_all").Do(ctx)
	if err != nil {
		return err
	}
	indices := res.IndicesByAlias(aliasName)
	if len(indices) > 0 {
		// Remove old aliases
		for _, indexName := range indices {
			svc.Remove(indexName, aliasName)
		}
	}
	// Add new alias
	svc.Add(indexName, aliasName)
	_, err = svc.Do(ctx)
	// Delete old indices
	if len(indices) > 0 {
		for _, indexName := range indices {
			ns.client.DeleteIndex(indexName).Do(ctx)
		}
	}
	return err
}

// UpdateAliasForType updates aliases
func (ns *EsIndexer) UpdateAliasForType(documentType string) {
	aliasName := ns.aliasNamePrefix + documentType
	indexName := ns.indexNamePrefix + documentType
	err := ns.UpdateAlias(aliasName, indexName)
	if err != nil {
		ns.log.Warn().Err(err).Str("aliasName", aliasName).Str("indexName", indexName).Msg("Error when updating alias")
	} else {
		ns.log.Info().Err(err).Str("aliasName", aliasName).Str("indexName", indexName).Msg("Updated alias")
	}
}

// OnSyncComplete is called when sync is finished catching up
func (ns *EsIndexer) OnSyncComplete() {
	if ns.reindexing {
		ns.reindexing = false
		ns.UpdateAliasForType("tx")
		ns.UpdateAliasForType("block")
	}
}

// Start setups the indexer
func (ns *EsIndexer) Start(grpcClient types.AergoRPCServiceClient, reindex bool) error {
	url := ns.esURL
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("http://%s", url)
	}
	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		return err
	}
	ns.client = client
	ns.grpcClient = grpcClient

	if reindex {
		ns.log.Warn().Msg("Reindexing database. Will sync from scratch and replace index aliases when caught up")
		ns.reindexing = true
	}

	ns.CreateIndexIfNotExists("tx")
	ns.CreateIndexIfNotExists("block")
	ns.UpdateLastBlockHeightFromDb()
	ns.log.Info().Uint64("lastBlockHeight", ns.lastBlockHeight).Msg("Started Elasticsearch Indexer")

	// Get new blocks from stream as they come in
	stream, err := grpcClient.ListBlockStream(context.Background(), &types.Empty{})
	if err != nil {
		return err
	}
	go func() {
		for {
			block, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				ns.log.Warn().Err(err).Msg("Failed to receive a block")
				return
			}
			ns.SyncBlock(block)
		}
	}()
	return nil
}

// Stop stops the indexer
func (ns *EsIndexer) Stop() {

}

// SyncBlock indexes new block after checking for skipped blocks and reorgs
func (ns *EsIndexer) SyncBlock(block *types.Block) {
	newHash := base58.Encode(block.Hash)
	newHeight := block.Header.BlockNo

	// Check out-of-sync cases
	if ns.lastBlockHeight == 0 && newHeight > 0 { // Initial sync
		// Add missing blocks asynchronously
		go ns.IndexBlocksInRange(0, newHeight-1)
	} else if newHeight > ns.lastBlockHeight+1 { // Skipped 1 or more blocks
		// Add missing blocks asynchronously
		go ns.IndexBlocksInRange(ns.lastBlockHeight+1, newHeight-1)
	} else if newHeight <= ns.lastBlockHeight { // Rewound 1 or more blocks
		// This needs to be syncronous, otherwise it may
		// delete the block we are just about to add
		ns.DeleteBlocksInRange(newHeight, ns.lastBlockHeight)
	}

	// Update state
	ns.lastBlockHash = newHash
	ns.lastBlockHeight = newHeight

	// Index new block
	go ns.IndexBlock(block)
}

// DeleteBlocksInRange deletes previously synced blocks in the range of [fromBlockheight, toBlockHeight]
func (ns *EsIndexer) DeleteBlocksInRange(fromBlockHeight uint64, toBlockHeight uint64) {
	ctx := context.Background()
	query := elastic.NewRangeQuery("no").From(fromBlockHeight).To(toBlockHeight)
	res, err := ns.client.DeleteByQuery().Index(ns.indexNamePrefix + "block").Query(query).Do(ctx)
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to delete blocks")
	}
	ns.log.Info().Int64("deleted", res.Deleted).Msg("Deleted blocks")
}

// GetBestBlockFromDb retrieves the current best block from the db
func (ns *EsIndexer) GetBestBlockFromDb() (EsBlock, error) {
	block := new(EsBlock)
	// Query best block
	ctx := context.Background()
	query := elastic.NewMatchAllQuery()
	res, err := ns.client.Search().Index(ns.indexNamePrefix+"block").Query(query).Sort("no", false).From(0).Size(1).Do(ctx)
	if err != nil {
		return *block, err
	}
	if res == nil || res.TotalHits() == 0 {
		return *block, errors.New("best block not found")
	}
	// Unmarshall block
	hit := res.Hits.Hits[0]
	if err := json.Unmarshal(*hit.Source, block); err != nil {
		return *block, err
	}
	block.BaseEsType = &BaseEsType{hit.Id}
	return *block, nil
}

// UpdateLastBlockHeightFromDb updates state from db
func (ns *EsIndexer) UpdateLastBlockHeightFromDb() {
	bestBlock, err := ns.GetBestBlockFromDb()
	if err != nil {
		return
	}
	ns.lastBlockHeight = bestBlock.BlockNo
	ns.lastBlockHash = bestBlock.id
}

// IndexBlock indexes one block
func (ns *EsIndexer) IndexBlock(block *types.Block) {
	ctx := context.Background()
	esBlock := ConvBlock(block)
	put, err := ns.client.Index().Index(ns.indexNamePrefix + "block").Type("block").Id(esBlock.id).BodyJson(esBlock).Do(ctx)
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to index block")
		return
	}
	ns.log.Info().Uint64("blockNo", block.Header.BlockNo).Str("blockHash", put.Id).Msg("Indexed block")

	if len(block.Body.Txs) > 0 {
		chunkSize := 5000
		ns.IndexTxs(block, block.Body.Txs, chunkSize)
	}
}

// IndexBlocksInRange indexes blocks in the range of [fromBlockheight, toBlockHeight]
func (ns *EsIndexer) IndexBlocksInRange(fromBlockHeight uint64, toBlockHeight uint64) {
	ctx := context.Background()
	channel := make(chan EsType)
	generator := func() error {
		defer close(channel)
		ns.log.Info().Msg(fmt.Sprintf("Indexing %d missing blocks [%d..%d]", (1 + toBlockHeight - fromBlockHeight), fromBlockHeight, toBlockHeight))
		for blockHeight := fromBlockHeight; blockHeight <= toBlockHeight; blockHeight++ {
			blockQuery := make([]byte, 8)
			binary.LittleEndian.PutUint64(blockQuery, uint64(blockHeight))
			block, err := ns.grpcClient.GetBlock(context.Background(), &types.SingleBytes{Value: blockQuery})
			if err != nil {
				ns.log.Warn().Uint64("blockHeight", blockHeight).Msg("Failed to get block")
				continue
			}
			if len(block.Body.Txs) > 0 {
				chunkSize := 5000
				ns.IndexTxs(block, block.Body.Txs, chunkSize)
			}
			d := ConvBlock(block)
			select {
			case channel <- d:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
	BulkIndexer(ctx, ns.log, ns.client, channel, generator, ns.indexNamePrefix+"block", "block", 5000)
	ns.OnSyncComplete()
}

// IndexTxs indexes a list of transactions in bulk
func (ns *EsIndexer) IndexTxs(block *types.Block, txs []*types.Tx, chunkSize int) {
	ctx := context.Background()
	channel := make(chan EsType)
	blockTs := time.Unix(0, block.Header.Timestamp)
	generator := func() error {
		defer close(channel)
		for _, tx := range txs {
			d := ConvTx(tx)
			d.Timestamp = blockTs
			d.BlockNo = block.Header.BlockNo
			select {
			case channel <- d:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
	BulkIndexer(ctx, ns.log, ns.client, channel, generator, ns.indexNamePrefix+"tx", "tx", 5000)
}
