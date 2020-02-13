#!/bin/sh
VERSION=${VERSION:=6.4.0}

#docker rm -f es_test
#docker rm -f aergo_test

echo "Starting elasticsearch"
docker run -d -p 9200:9200 --name es_test -e "http.host=0.0.0.0" -e "transport.host=127.0.0.1" -e "bootstrap.memory_lock=true" -e "ES_JAVA_OPTS=-Xms1g -Xmx1g" docker.elastic.co/elasticsearch/elasticsearch-oss:$VERSION elasticsearch -Enetwork.host=_local_,_site_ -Enetwork.publish_host=_local_

echo "Starting mariadb"
#docker run -d -p 3306:3306 --name mariadb_test -e MYSQL_ROOT_PASSWORD=my-secret-pw mariadb:10.4-bionic

echo "Starting aergosvr"
#docker run -d -p 7845:7845 --name aergo_test aergo/node:1.3 aergosvr --config /aergo/testmode.toml
docker run -d -p 7845:7845 --name aergo_test aergo/node:2.0 aergosvr --config /aergo/testmode.toml

#docker run -d -p 7845:7845 --name aergo_test aergo/node
echo "Starting indexer"
sleep 3

#AERGO_URL="localhost:7845"
AERGO_URL="mainnet-api.aergo.io:7845"
ES_URL="http://localhost:9200"
MARIADB_URL="root:my-secret-pw@tcp(localhost:3306)/test"
CHAIN_PREFIX="chain_"
#SYNC_FROM=19915001
SYNC_FROM=21257091
#SYNC_TO=19916000
SYNC_TO=30000000

#time ./bin/indexer -A $AERGO_URL --dbtype mariadb --dburl $MARIADB_URL --prefix $CHAIN_PREFIX --from $SYNC_FROM --to $SYNC_TO --exit-on-complete --reindex

./bin/indexer -A $AERGO_URL --dbtype elastic --dburl $ES_URL --prefix $CHAIN_PREFIX --from $SYNC_FROM --to $SYNC_TO --conflict 10

# time ../aergo-esindexer/bin/esindexer -A $AERGO_URL -E $ES_URL --prefix old_$CHAIN_PREFIX --exit-on-complete --reindex 

#./bin/indexer -H localhost -p 7845 --dbtype mariadb --dburl "root:my-secret-pw@tcp(localhost:3306)/test" --prefix chain_
#./bin/indexer -H localhost -p 7845 --dburl http://localhost:9200 --prefix chain_
# ./bin/indexer -H localhost -p 7845 --dburl http://localhost:9200 --prefix chain_ --reindex
#./bin/indexer -H localhost -p 7845 --dburl http://localhost:9200 -A testnet.aergo.io:7845 --reindex --prefix chain_
# docker rm -f es_test
# docker rm -f aergo_test
# docker rm -f mariadb_test