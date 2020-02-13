module github.com/aergoio/aergo-indexer

go 1.12

require (
	github.com/aergoio/aergo-lib v0.0.0-20181112043831-5ddbd01605af
	github.com/anaskhan96/base58check v0.0.0-20171020155424-fcff33ba49dd
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.3.2
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/graup/es-distributed-lock v0.0.1
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/mr-tron/base58 v0.0.0-20181030092856-c8897612421d
	github.com/olivere/elastic v6.2.27+incompatible
	github.com/rs/zerolog v1.13.0 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/spf13/viper v1.2.1 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.0.0-20191002035440-2ec189313ef0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20191220142924-d4481acd189f // indirect
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c // indirect
	google.golang.org/grpc v1.24.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.3 // indirect
)

// For local dev
// replace github.com/graup/es-distributed-lock v0.0.1 => /Users/paulgrau/go/src/github.com/graup/es-distributed-lock
