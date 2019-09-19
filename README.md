# Aergo Metadata Indexer

This is a go program that connects to aergo server over RPC and synchronizes blockchain metadata with a database. It currently supports Elastic Search and MySQL/MariaDB.

This creates the indices `block`, `tx`, and `name` (with a prefix). These are actually aliases that point to the latest version of the data.
Check [esindexer/types.go](./esindexer/types.go) for the exact index mappings.

## Indexed data

Blocks
```
Field    Type        Comment
_id      string      block hash
ts       timestamp   block creation timestamp
no       uint64      block number
txs      uint        number of transactions
```

Transaction
```
Field    Type        Comment
_id      string      tx hash
ts       timestamp   block creation timestamp
blockno  uint64      block number
from     string
to       string
amount   string
type     string      "0" or "1"
payload0 byte        first byte of payload
category string      user-friendly category
```

Names
```
Field    Type        Comment
_id      string      name + tx hash
name     string
address  string
blockno  uint64      block in which name was updated
tx       string      tx in which name was updated
```

## Build

    go get github.com/aergoio/aergo-indexer
    cd $GOPATH/src/github.com/aergoio/aergo-indexer
    make

## Usage

    ./bin/indexer -H localhost -p 7845 --dburl http://localhost:9200 --prefix chain_

You can use the `--prefix` parameter and multiple instances of this program to sync several blockchains with one database.

Instead of setting host and port of the aergo server separately, you can also pass them at once with `-A localhost:7845`.

To reindex (starting from scratch):

    ./bin/indexer --reindex

When reindexing, this creates new indices to sync the blockchain from scratch.
After catching up, the aliases are replaced with the new data and the old indices removed.
This means the old data can still be accessed until the sync is complete.

## Build and run using Docker

    docker build -t aergo/indexer .
    docker run aergo/indexer indexer -A ip:7845 -E ip:9200 --prefix chain_