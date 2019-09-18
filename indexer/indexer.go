package indexer

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/aergoio/aergo-indexer/indexer/db"
	doc "github.com/aergoio/aergo-indexer/indexer/documents"
	"github.com/aergoio/aergo-indexer/types"
	"github.com/aergoio/aergo-lib/log"
	"github.com/mr-tron/base58/base58"
)

// Indexer hold all state information
type Indexer struct {
	db              db.DbController
	grpcClient      types.AergoRPCServiceClient
	aliasNamePrefix string
	indexNamePrefix string
	lastBlockHeight uint64
	lastBlockHash   string
	log             *log.Logger
	reindexing      bool
	exitOnComplete  bool
	State           string
	BulkState       string
	stream          types.AergoRPCService_ListBlockStreamClient
	startFrom       int64
	stopAt          int64
}

// NewIndexer creates new Indexer instance
func NewIndexer(logger *log.Logger, dbType string, dbURL string, namePrefix string) (*Indexer, error) {
	aliasNamePrefix := namePrefix
	var dbController db.DbController
	var err error
	switch dbType {
	case "es":
		dbController, err = db.NewElasticsearchDbController(dbURL)
	case "mariadb":
		dbController, err = db.NewMariaDbController(dbURL)
	default:
		return nil, fmt.Errorf("Invalid database type: %s", dbType)
	}
	if err != nil {
		return nil, err
	}
	logger.Info().Str("dbType", dbType).Str("dbURL", dbURL).Msg("Initialized database connection")
	svc := &Indexer{
		db:              dbController,
		aliasNamePrefix: aliasNamePrefix,
		indexNamePrefix: generateIndexPrefix(aliasNamePrefix),
		lastBlockHeight: 0,
		lastBlockHash:   "",
		State:           "booting",
		BulkState:       "finished",
		log:             logger,
		reindexing:      false,
		exitOnComplete:  false,
		startFrom:       0,
		stopAt:          -1,
	}
	return svc, nil
}

func generateIndexPrefix(aliasNamePrefix string) string {
	return fmt.Sprintf("%s%s_", aliasNamePrefix, time.Now().UTC().Format("2006-01-02_15-04-05"))
}

// CreateIndexIfNotExists creates the indices and aliases in ES
func (ns *Indexer) CreateIndexIfNotExists(documentType string) {
	initialized := true
	aliasName := ns.aliasNamePrefix + documentType
	// Check for existing index to find out current indexNamePrefix
	if !ns.reindexing {
		exists, indexNamePrefix, err := ns.db.GetExistingIndexPrefix(aliasName, documentType)
		if err != nil {
			ns.log.Error().Err(err).Msg("Error when checking for alias")
		}
		if exists {
			ns.log.Info().Str("aliasName", aliasName).Str("indexNamePrefix", indexNamePrefix).Msg("Alias found")
			ns.indexNamePrefix = indexNamePrefix
		} else {
			initialized = false
			ns.reindexing = false
		}
	}
	// Create new index
	if ns.reindexing || !initialized {
		indexName := ns.indexNamePrefix + documentType

		err := ns.db.CreateIndex(indexName, documentType)
		if err != nil {
			ns.log.Error().Err(err).Str("indexName", indexName).Msg("Error when creating index")
		} else {
			ns.log.Info().Str("indexName", indexName).Msg("Created index")
		}
		// Update alias, only when initializing and not reindexing
		if !ns.reindexing {
			err = ns.db.UpdateAlias(aliasName, indexName)
			if err != nil {
				ns.log.Error().Err(err).Str("aliasName", aliasName).Str("indexName", indexName).Msg("Error when updating alias")
			} else {
				ns.log.Info().Str("aliasName", aliasName).Str("indexName", indexName).Msg("Updated alias")
			}
		}
	}
}

// UpdateAliasForType updates aliases
func (ns *Indexer) UpdateAliasForType(documentType string) {
	aliasName := ns.aliasNamePrefix + documentType
	indexName := ns.indexNamePrefix + documentType
	err := ns.db.UpdateAlias(aliasName, indexName)
	if err != nil {
		ns.log.Warn().Err(err).Str("aliasName", aliasName).Str("indexName", indexName).Msg("Error when updating alias")
	} else {
		ns.log.Info().Err(err).Str("aliasName", aliasName).Str("indexName", indexName).Msg("Updated alias")
	}
}

// OnSyncComplete is called when sync is finished catching up
func (ns *Indexer) OnSyncComplete() {
	if ns.reindexing {
		ns.reindexing = false
		ns.UpdateAliasForType("tx")
		ns.UpdateAliasForType("block")
		ns.UpdateAliasForType("name")
	}
	ns.log.Info().Msg("Initial sync complete")
	if ns.exitOnComplete {
		ns.Stop()
	}
}

// Start setups the indexer
func (ns *Indexer) Start(grpcClient types.AergoRPCServiceClient, reindex bool, exitOnComplete bool, startFrom int64, stopAt int64) error {
	ns.grpcClient = grpcClient

	if reindex {
		ns.log.Warn().Msg("Reindexing database. Will sync from scratch and replace index aliases when caught up")
		ns.reindexing = true
		ns.exitOnComplete = exitOnComplete
	}

	ns.CreateIndexIfNotExists("tx")
	ns.CreateIndexIfNotExists("block")
	ns.CreateIndexIfNotExists("name")
	ns.UpdateLastBlockHeightFromDb()
	ns.log.Info().Uint64("last block height", ns.lastBlockHeight).Msg("Started Indexer")

	ns.startFrom = startFrom
	ns.stopAt = stopAt
	if startFrom != 0 || stopAt != -1 {
		ns.log.Info().Int64("startFrom", startFrom).Int64("stopAt", stopAt).Msg("Only index block number range")
	}

	go ns.CheckConsistency()

	if ns.reindexing {
		// Don't wait for sync to start when blockchain is booting from genesis
		nodeBlockheight, err := ns.GetNodeBlockHeight()
		if err != nil {
			ns.log.Warn().Err(err).Msg("Failed to query node's block height")
		} else {
			if nodeBlockheight == 0 {
				ns.OnSyncComplete()
			}
		}
	}

	return ns.StartStream()
}

// StartStream starts the block stream and calls SyncBlock
func (ns *Indexer) StartStream() error {
	var err error
	ns.stream, err = ns.grpcClient.ListBlockStream(context.Background(), &types.Empty{})
	if err != nil {
		return err
	}
	ns.State = "running"
	go func() {
		for {
			if ns.State == "stopped" {
				return
			}
			block, err := ns.stream.Recv()
			if err == io.EOF {
				ns.log.Warn().Msg("Stream ended")
				ns.RestartStream()
				return
			}
			if err != nil {
				ns.log.Warn().Err(err).Msg("Failed to receive a block")
				ns.RestartStream()
				return
			}
			ns.SyncBlock(block)
		}
	}()
	return nil
}

// RestartStream restarts the streem after a few seconds and keeps trying to start
func (ns *Indexer) RestartStream() {
	if ns.stream != nil {
		ns.stream.CloseSend()
		ns.stream = nil
	}
	ns.log.Info().Msg("Restarting stream in 5 seconds")
	ns.State = "restarting"
	time.Sleep(5 * time.Second)
	err := ns.StartStream()
	if err != nil {
		ns.log.Error().Err(err).Msg("Failed to restart stream")
		ns.RestartStream()
	}
}

// Stop stops the indexer
func (ns *Indexer) Stop() {
	if ns.stream != nil {
		ns.stream.CloseSend()
		ns.stream = nil
		ns.State = "stopped"
	}
}

// SyncBlock indexes new block after checking for skipped blocks and reorgs
func (ns *Indexer) SyncBlock(block *types.Block) {
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

	// Check specified sync range
	if newHeight < uint64(ns.startFrom) {
		ns.log.Info().Uint64("blockNumber", newHeight).Int64("startFrom", ns.startFrom).Msg("Skipping block before specified sync range")
		return
	}
	if ns.stopAt != -1 && newHeight > uint64(ns.stopAt) {
		ns.log.Info().Uint64("blockNumber", newHeight).Int64("stopAt", ns.stopAt).Msg("Reached end of specified sync range")
		//ns.OnSyncComplete()
		ns.Stop()
		return
	}

	// Index new block
	go ns.IndexBlock(block)
}

// GetBestBlockFromDb retrieves the current best block from the db
func (ns *Indexer) GetBestBlockFromDb() (*doc.EsBlock, error) {
	block, err := ns.db.SelectOne(db.QueryParams{
		IndexName: ns.indexNamePrefix + "block",
		SortField: "no",
		SortAsc:   false,
	}, func() doc.DocType {
		block := new(doc.EsBlock)
		return block
	})

	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, errors.New("best block not found")
	}
	return block.(*doc.EsBlock), nil
}

// UpdateLastBlockHeightFromDb updates state from db
func (ns *Indexer) UpdateLastBlockHeightFromDb() {
	bestBlock, err := ns.GetBestBlockFromDb()
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to update best block")
		return
	}
	ns.lastBlockHeight = bestBlock.BlockNo
	ns.lastBlockHash = bestBlock.GetID()
}

// GetNodeBlockHeight updates state from db
func (ns *Indexer) GetNodeBlockHeight() (uint64, error) {
	blockchain, err := ns.grpcClient.Blockchain(context.Background(), &types.Empty{})
	if err != nil {
		return 0, err
	}
	return blockchain.BestHeight, nil
}

// IndexBlock indexes one block
func (ns *Indexer) IndexBlock(block *types.Block) {
	ctx := context.Background()
	blockDocument := ns.ConvBlock(block)
	_, err := ns.db.Insert(blockDocument, db.UpdateParams{IndexName: ns.indexNamePrefix + "block", TypeName: "block"})
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to index block")
		return
	}

	// Index one block's transactions
	if len(block.Body.Txs) > 0 {
		txChannel := make(chan doc.DocType)
		nameChannel := make(chan doc.DocType)
		done := make(chan struct{})

		waitForNames := func() error {
			defer close(nameChannel)
			<-done
			return nil
		}
		go BulkIndexer(ctx, ns.log, ns.db, nameChannel, waitForNames, ns.indexNamePrefix+"name", "name", 2500, true)

		generator := func() error {
			defer close(txChannel)
			defer close(done)
			ns.IndexTxs(block, block.Body.Txs, txChannel, nameChannel)
			return nil
		}
		BulkIndexer(ctx, ns.log, ns.db, txChannel, generator, ns.indexNamePrefix+"tx", "tx", 2000, false)
	}

	ns.log.Info().Uint64("no", block.Header.BlockNo).Int("txs", len(block.Body.Txs)).Str("hash", blockDocument.GetID()).Msg("Indexed block")
}

// IndexBlocksInRange indexes blocks in the range of [fromBlockheight, toBlockHeight]
func (ns *Indexer) IndexBlocksInRange(fromBlockHeight uint64, toBlockHeight uint64) {
	ns.BulkState = "running"
	ctx := context.Background()
	channel := make(chan doc.DocType, 1000)
	done := make(chan struct{})
	txChannel := make(chan doc.DocType, 20000)
	nameChannel := make(chan doc.DocType, 5000)

	if fromBlockHeight < uint64(ns.startFrom) {
		fromBlockHeight = uint64(ns.startFrom)
	}
	if toBlockHeight > uint64(ns.stopAt) {
		toBlockHeight = uint64(ns.stopAt)
	}

	var wg sync.WaitGroup

	waitForTx := func() error {
		defer close(txChannel)
		<-done
		return nil
	}
	go func() {
		defer wg.Done()
		BulkIndexer(ctx, ns.log, ns.db, txChannel, waitForTx, ns.indexNamePrefix+"tx", "tx", 2000, false)
	}()
	wg.Add(1)

	waitForNames := func() error {
		defer close(nameChannel)
		<-done
		return nil
	}
	go func() {
		defer wg.Done()
		BulkIndexer(ctx, ns.log, ns.db, nameChannel, waitForNames, ns.indexNamePrefix+"name", "name", 2500, true)
	}()
	wg.Add(1)

	generator := func() error {
		defer close(channel)
		defer close(done)
		ns.log.Info().Msg(fmt.Sprintf("Indexing %d missing blocks [%d..%d]", (1 + toBlockHeight - fromBlockHeight), fromBlockHeight, toBlockHeight))
		for blockHeight := fromBlockHeight; blockHeight <= toBlockHeight; blockHeight++ {
			blockQuery := make([]byte, 8)
			binary.LittleEndian.PutUint64(blockQuery, uint64(blockHeight))
			block, err := ns.grpcClient.GetBlock(context.Background(), &types.SingleBytes{Value: blockQuery})
			if err != nil {
				ns.log.Warn().Uint64("blockHeight", blockHeight).Err(err).Msg("Failed to get block")
				continue
			}
			if len(block.Body.Txs) > 0 {
				ns.IndexTxs(block, block.Body.Txs, txChannel, nameChannel)
			}
			d := ns.ConvBlock(block)
			select {
			case channel <- d:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
	BulkIndexer(ctx, ns.log, ns.db, channel, generator, ns.indexNamePrefix+"block", "block", 500, false)

	// Wait for tx and name goroutines
	wg.Wait()
	ns.BulkState = "finished"
	ns.OnSyncComplete()
}

// IndexTxs indexes a list of transactions in bulk
func (ns *Indexer) IndexTxs(block *types.Block, txs []*types.Tx, channel chan doc.DocType, nameChannel chan doc.DocType) {
	// This simply pushes all Txs to the channel to be consumed elsewhere
	blockTs := time.Unix(0, block.Header.Timestamp)
	for _, tx := range txs {
		d := ns.ConvTx(tx, block.Header.BlockNo)
		d.Timestamp = blockTs
		d.BlockNo = block.Header.BlockNo

		// Add tx to channel
		channel <- d

		// Process name transactions
		if tx.GetBody().GetType() == types.TxType_GOVERNANCE && string(tx.GetBody().GetRecipient()) == "aergo.name" {
			nameDoc := ns.ConvNameTx(tx, d.BlockNo)
			nameDoc.UpdateBlock = d.BlockNo
			nameChannel <- nameDoc
		}
	}
}

func (ns *Indexer) deleteTypeByQuery(typeName string, rangeQuery db.IntegerRangeQuery) {
	deleted, err := ns.db.Delete(db.QueryParams{
		IndexName:    ns.indexNamePrefix + typeName,
		IntegerRange: &rangeQuery,
	})
	if err != nil {
		ns.log.Warn().Err(err).Str("typeName", typeName).Msg("Failed to delete documents")
	} else {
		ns.log.Info().Uint64("deleted", deleted).Str("typeName", typeName).Msg("Deleted documents")
	}
}

// DeleteBlocksInRange deletes previously synced blocks and their txs and names in the range of [fromBlockheight, toBlockHeight]
func (ns *Indexer) DeleteBlocksInRange(fromBlockHeight uint64, toBlockHeight uint64) {
	ns.log.Info().Msg(fmt.Sprintf("Rolling back %d blocks [%d..%d]", (1 + toBlockHeight - fromBlockHeight), fromBlockHeight, toBlockHeight))
	ns.deleteTypeByQuery("block", db.IntegerRangeQuery{Field: "no", Min: fromBlockHeight, Max: toBlockHeight})
	ns.deleteTypeByQuery("tx", db.IntegerRangeQuery{Field: "blockno", Min: fromBlockHeight, Max: toBlockHeight})
	ns.deleteTypeByQuery("name", db.IntegerRangeQuery{Field: "blockno", Min: fromBlockHeight, Max: toBlockHeight})
}
