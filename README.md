# Aergo Metadata Indexer for Elasticsearch

This is a go program that connects to aergo server over RPC and synchronizes blockchain metadata with an Elasticsearch cluster.

This creates the indices `chain_block` and `chain_tx`. These are actually aliases that point to the latest version of the data.

## Build

You need Glide to install dependencies.

    make all

## Usage

    ./bin/esindexer -H localhost -p 7845 --esurl http://localhost:9200 --prefix chain_

You can use the ``--prefix` parameter and multiple instances of this program to sync several blockchains with one database.

Instead of setting host and port of the aergo separately, you can also pass them at once with `-A localhost:7845`.

To reindex (starting from scratch):

    ./bin/esindexer --reindex

When reindexing, this creates new indices to sync the blockchain from scratch.
After catching up the first time, the aliases are replaced with the new data and the old indices removed.
This means the old data can still be accessed until the sync is complete.