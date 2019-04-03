package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aergoio/aergo-esindexer/esindexer"
	"github.com/aergoio/aergo-esindexer/types"
	"github.com/aergoio/aergo-lib/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	rootCmd = &cobra.Command{
		Use:   "esindexer",
		Short: "Aergo Elasticsearch Indexer",
		Long:  "Aergo Metadata Indexer for Elasticsearch",
		Run:   rootRun,
	}
	reindexingMode  bool
	host            string
	port            int32
	esURL           string
	indexNamePrefix string
	aergoAddress    string

	logger *log.Logger

	client  types.AergoRPCServiceClient
	indexer *esindexer.EsIndexer
)

func init() {
	fs := rootCmd.PersistentFlags()
	fs.BoolVar(&reindexingMode, "reindex", false, "reindex blocks from genesis and swap index after catching up")
	fs.StringVarP(&host, "host", "H", "localhost", "host address of aergo server")
	fs.Int32VarP(&port, "port", "p", 7845, "port number of aergo server")
	fs.StringVarP(&aergoAddress, "aergo", "A", "", "host and port of aergo server. Alternative to setting host and port separately.")
	fs.StringVarP(&esURL, "esurl", "E", "http://127.0.0.1:9200", "URL of elasticsearch server")
	fs.StringVarP(&indexNamePrefix, "prefix", "X", "chain_", "prefix used for index names")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootRun(cmd *cobra.Command, args []string) {
	logger = log.NewLogger("esindexer")
	logger.Info().Msg("Starting")

	indexer = esindexer.NewEsIndexer(logger, esURL, indexNamePrefix)
	client = waitForClient(getServerAddress())

	err := indexer.Start(client, reindexingMode)
	if err != nil {
		logger.Warn().Err(err).Str("esURL", esURL).Msg("Could not start elasticsearch indexer")
		return
	}

	handleKillSig(func() {
		indexer.Stop()
	}, logger)

	for {
		time.Sleep(time.Minute)
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
		conn, err = grpc.Dial(serverAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
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
