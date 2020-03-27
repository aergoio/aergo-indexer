package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	indx "github.com/aergoio/aergo-indexer/indexer"
	"github.com/aergoio/aergo-indexer/types"
	"github.com/aergoio/aergo-lib/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	rootCmd = &cobra.Command{
		Use:   "indexer",
		Short: "Aergo Indexer",
		Long:  "Aergo Metadata Indexer",
		Run:   rootRun,
	}
	reindexingMode  bool
	exitOnComplete  bool
	host            string
	port            int32
	dbURL           string
	dbType          string
	indexNamePrefix string
	aergoAddress    string
	startFrom       int32
	stopAt          int32
	idleOnConflict  int32

	logger *log.Logger

	client  types.AergoRPCServiceClient
	indexer *indx.Indexer
)

func init() {
	fs := rootCmd.PersistentFlags()
	fs.BoolVar(&reindexingMode, "reindex", false, "reindex blocks from genesis and swap index after catching up")
	fs.BoolVar(&exitOnComplete, "exit-on-complete", false, "exit when reindexing sync completes for the first time")
	fs.StringVarP(&host, "host", "H", "localhost", "host address of aergo server")
	fs.Int32VarP(&port, "port", "p", 7845, "port number of aergo server")
	fs.StringVarP(&aergoAddress, "aergo", "A", "", "host and port of aergo server. Alternative to setting host and port separately.")
	fs.StringVarP(&dbURL, "dburl", "E", "http://localhost:9200", "Database URL")
	fs.StringVarP(&dbType, "dbtype", "T", "elastic", "Type of database used (elastic, mariadb)")
	fs.StringVarP(&indexNamePrefix, "prefix", "X", "chain_", "prefix used for index names")
	fs.Int32VarP(&startFrom, "from", "", 0, "start syncing from this block number")
	fs.Int32VarP(&stopAt, "to", "", -1, "stop syncing at this block number")
	fs.Int32VarP(&idleOnConflict, "conflict", "", 0, "time to idle when a conflict occurs (in seconds). Use this for optimistic concurrency. Elasticsearch only")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(cmd *cobra.Command, args []string) {
	logger = log.NewLogger("indexer")

	logger.Info().Msg("Starting indexer...")

	indexer, err := indx.NewIndexer(logger, dbType, dbURL, indexNamePrefix)
	if err != nil {
		logger.Warn().Err(err).Str("dbURL", dbURL).Msg("Could not start indexer")
		return
	}
	client = waitForClient(getServerAddress())

	err = indexer.Start(client, reindexingMode, exitOnComplete, int64(startFrom), int64(stopAt), idleOnConflict)
	if err != nil {
		logger.Warn().Err(err).Str("dbURL", dbURL).Msg("Could not start indexer")
		return
	}

	handleKillSig(func() {
		indexer.Stop()
	}, logger)

	for {
		if exitOnComplete {
			if indexer.State == "stopped" && indexer.BulkState == "finished" {
				break
			}
			time.Sleep(time.Second)
		} else {
			time.Sleep(time.Minute)
		}
	}
}

func getServerAddress() string {
	if len(aergoAddress) > 0 {
		return aergoAddress
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func waitForClient(serverAddr string) types.AergoRPCServiceClient {
	var conn *grpc.ClientConn
	var err error
	for {
		ctx := context.Background()
		maxMsgSize := 1024 * 1024 * 10 // 10mb
		conn, err = grpc.DialContext(ctx, serverAddr,
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(5*time.Second),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)),
		)
		if err == nil && conn != nil {
			break
		}
		logger.Info().Str("serverAddr", serverAddr).Err(err).Msg("Could not connect to aergo server, retrying")
		time.Sleep(time.Second)
	}
	logger.Info().Str("serverAddr", serverAddr).Msg("Connected to aergo server")
	return types.NewAergoRPCServiceClient(conn)
}

func handleKillSig(handler func(), logger *log.Logger) {
	sigChannel := make(chan os.Signal, 1)

	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for signal := range sigChannel {
			logger.Info().Msgf("Receive signal %s, Shutting down...", signal)
			handler()
			os.Exit(1)
		}
	}()
}
