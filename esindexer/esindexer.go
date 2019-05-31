package esindexer

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aergoio/aergo-esindexer/types"
	"github.com/aergoio/aergo-lib/log"
	"github.com/mr-tron/base58/base58"
	"github.com/olivere/elastic"
)

// EsIndexer hold all state information
type EsIndexer struct {
	client          *elastic.Client
	grpcClient      types.AergoRPCServiceClient
	aliasNamePrefix string
	indexNamePrefix string
	lastBlockHeight uint64
	lastBlockHash   string
	log             *log.Logger
	reindexing      bool
	exitOnComplete  bool
	State           string
	esURL           string
	stream          types.AergoRPCService_ListBlockStreamClient
	nameCache       map[string]string
}

// NewEsIndexer creates new EsIndexer instance
func NewEsIndexer(logger *log.Logger, esURL string, namePrefix string) *EsIndexer {
	aliasNamePrefix := namePrefix
	svc := &EsIndexer{
		esURL:           esURL,
		aliasNamePrefix: aliasNamePrefix,
		indexNamePrefix: generateIndexPrefix(aliasNamePrefix),
		lastBlockHeight: 0,
		lastBlockHash:   "",
		State:           "booting",
		log:             logger,
		reindexing:      false,
		exitOnComplete:  false,
		nameCache:       map[string]string{},
	}
	return svc
}

func generateIndexPrefix(aliasNamePrefix string) string {
	return fmt.Sprintf("%s%s_", aliasNamePrefix, time.Now().UTC().Format("2006-01-02_15-04-05"))
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
		createIndex, err := ns.client.CreateIndex(indexName).BodyString(mappings[documentType]).Do(ctx)
		if err != nil || !createIndex.Acknowledged {
			ns.log.Warn().Err(err).Str("indexName", indexName).Msg("Error when creating index")
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
	ns.log.Info().Msg("Sync complete")
	if ns.reindexing {
		ns.reindexing = false
		ns.UpdateAliasForType("tx")
		ns.UpdateAliasForType("block")
		ns.UpdateAliasForType("name")
		if ns.exitOnComplete {
			ns.Stop()
		}
	}
}

// Start setups the indexer
func (ns *EsIndexer) Start(grpcClient types.AergoRPCServiceClient, reindex bool, exitOnComplete bool) error {
	url := ns.esURL
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("http://%s", url)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	client, err := elastic.NewClient(
		elastic.SetHttpClient(httpClient),
		elastic.SetURL(url),
		elastic.SetHealthcheckTimeoutStartup(30*time.Second),
		elastic.SetSniff(false),
	)
	if err != nil {
		return err
	}
	ns.client = client
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
	ns.log.Info().Uint64("lastBlockHeight", ns.lastBlockHeight).Msg("Started Elasticsearch Indexer")

	go ns.CheckConsistency()

	if ns.reindexing {
		// Don't wait for sync to start when blockchain is booting
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

type esBlockNo struct {
	BlockNo uint64 `json:"no"`
}

// CheckConsistency gets all block numbers from 0 to ns.lastBlockHeight in order and checks for "holes"
func (ns *EsIndexer) CheckConsistency() {
	ctx := context.Background()

	count, err := ns.client.Count(ns.indexNamePrefix + "block").Do(ctx)
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to query block count")
		return
	}
	ns.log.Info().Int64("total indexed", count).Uint64("expected", ns.lastBlockHeight+1).Msg("Checking consistency")
	if uint64(count) >= ns.lastBlockHeight+1 {
		return
	}

	prevBlockNo := uint64(0)
	missingBlocks := uint64(0)

	fsc := elastic.NewFetchSourceContext(true).Include("no")
	scroll := ns.client.Scroll(ns.indexNamePrefix+"block").Type("block").Size(10000).Sort("no", true).FetchSourceContext(fsc)
	for {
		results, err := scroll.Do(ctx)
		if err == io.EOF {
			break // all results retrieved
		}
		if err != nil {
			ns.log.Warn().Err(err).Msg("Failed to query block numbers")
			break // something went wrong
		}

		for _, hit := range results.Hits.Hits {
			blockNo := new(esBlockNo)
			if err := json.Unmarshal(*hit.Source, blockNo); err != nil {
				ns.log.Warn().Err(err).Msg("Failed to unmarshal blockNo source")
				continue
			}
			if blockNo.BlockNo > prevBlockNo+1 {
				missingBlocks = missingBlocks + (blockNo.BlockNo - prevBlockNo - 1)
				ns.IndexBlocksInRange(prevBlockNo+1, blockNo.BlockNo-1)
			}
			prevBlockNo = blockNo.BlockNo
		}
	}

	ns.log.Info().Uint64("missing", missingBlocks).Msg("Done with consistency check")
}

// StartStream starts the block stream and calls SyncBlock
func (ns *EsIndexer) StartStream() error {
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
func (ns *EsIndexer) RestartStream() {
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
func (ns *EsIndexer) Stop() {
	if ns.stream != nil {
		ns.stream.CloseSend()
		ns.stream = nil
		ns.State = "stopped"
	}
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

// GetNodeBlockHeight updates state from db
func (ns *EsIndexer) GetNodeBlockHeight() (uint64, error) {
	blockchain, err := ns.grpcClient.Blockchain(context.Background(), &types.Empty{})
	if err != nil {
		return 0, err
	}
	return blockchain.BestHeight, nil
}

// IndexBlock indexes one block
func (ns *EsIndexer) IndexBlock(block *types.Block) {
	ctx := context.Background()
	esBlock := ns.ConvBlock(block)
	put, err := ns.client.Index().Index(ns.indexNamePrefix + "block").Type("block").Id(esBlock.id).BodyJson(esBlock).Do(ctx)
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to index block")
		return
	}

	if len(block.Body.Txs) > 0 {
		txChannel := make(chan EsType)
		nameChannel := make(chan EsType)
		done := make(chan struct{})
		generator := func() error {
			defer close(txChannel)
			defer close(done)
			ns.IndexTxs(block, block.Body.Txs, txChannel, nameChannel)
			return nil
		}

		waitForNames := func() error {
			defer close(nameChannel)
			<-done
			return nil
		}
		go BulkIndexer(ctx, ns.log, ns.client, nameChannel, waitForNames, ns.indexNamePrefix+"name", "name", 2500, true)

		BulkIndexer(ctx, ns.log, ns.client, txChannel, generator, ns.indexNamePrefix+"tx", "tx", 10000, false)
	}

	ns.log.Info().Uint64("blockNo", block.Header.BlockNo).Int("txs", len(block.Body.Txs)).Str("blockHash", put.Id).Msg("Indexed block")
}

// IndexBlocksInRange indexes blocks in the range of [fromBlockheight, toBlockHeight]
func (ns *EsIndexer) IndexBlocksInRange(fromBlockHeight uint64, toBlockHeight uint64) {
	ctx := context.Background()
	channel := make(chan EsType, 1000)
	done := make(chan struct{})
	txChannel := make(chan EsType, 20000)
	nameChannel := make(chan EsType, 5000)
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

	waitForTx := func() error {
		defer close(txChannel)
		<-done
		return nil
	}
	go BulkIndexer(ctx, ns.log, ns.client, txChannel, waitForTx, ns.indexNamePrefix+"tx", "tx", 10000, false)

	waitForNames := func() error {
		defer close(nameChannel)
		<-done
		return nil
	}
	go BulkIndexer(ctx, ns.log, ns.client, nameChannel, waitForNames, ns.indexNamePrefix+"name", "name", 2500, true)

	BulkIndexer(ctx, ns.log, ns.client, channel, generator, ns.indexNamePrefix+"block", "block", 500, false)

	ns.OnSyncComplete()
}

// IndexTxs indexes a list of transactions in bulk
func (ns *EsIndexer) IndexTxs(block *types.Block, txs []*types.Tx, channel chan EsType, nameChannel chan EsType) {
	var nameCacheUpdates []EsName
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
			nameCacheUpdates = append(nameCacheUpdates, nameDoc)
			nameChannel <- nameDoc
		}
	}
	// Update name cache AFTER all tx of this block are processed
	for _, nameDoc := range nameCacheUpdates {
		ns.nameCache[nameDoc.Name] = nameDoc.Address
	}
}

// DeleteBlocksInRange deletes previously synced blocks and their txs in the range of [fromBlockheight, toBlockHeight]
func (ns *EsIndexer) DeleteBlocksInRange(fromBlockHeight uint64, toBlockHeight uint64) {
	ctx := context.Background()
	ns.log.Info().Msg(fmt.Sprintf("Rolling back %d blocks [%d..%d]", (1 + toBlockHeight - fromBlockHeight), fromBlockHeight, toBlockHeight))
	// Delete blocks
	query := elastic.NewRangeQuery("no").From(fromBlockHeight).To(toBlockHeight)
	res, err := ns.client.DeleteByQuery().Index(ns.indexNamePrefix + "block").Query(query).Do(ctx)
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to delete blocks")
	} else {
		ns.log.Info().Int64("deleted", res.Deleted).Msg("Deleted blocks")
	}
	// Delete tx of blocks
	query = elastic.NewRangeQuery("blockno").From(fromBlockHeight).To(toBlockHeight)
	res, err = ns.client.DeleteByQuery().Index(ns.indexNamePrefix + "tx").Query(query).Do(ctx)
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to delete tx")
	} else {
		ns.log.Info().Int64("deleted", res.Deleted).Msg("Deleted tx")
	}
	// Delete invalidated name entries
	query = elastic.NewRangeQuery("blockno").From(fromBlockHeight).To(toBlockHeight)
	res, err = ns.client.DeleteByQuery().Index(ns.indexNamePrefix + "name").Query(query).Do(ctx)
	if err != nil {
		ns.log.Warn().Err(err).Msg("Failed to delete names")
	} else {
		ns.log.Info().Int64("deleted", res.Deleted).Msg("Deleted names")
	}
}
