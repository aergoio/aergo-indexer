#!/bin/sh
VERSION=${VERSION:=6.4.0}

#docker rm -f es_test

echo "Starting elasticsearch"
docker run -d -p 9200:9200 --name es_test -e "http.host=0.0.0.0" -e "transport.host=127.0.0.1" -e "bootstrap.memory_lock=true" -e "ES_JAVA_OPTS=-Xms1g -Xmx1g" docker.elastic.co/elasticsearch/elasticsearch-oss:$VERSION elasticsearch -Enetwork.host=_local_,_site_ -Enetwork.publish_host=_local_

#docker run -d -p 7845:7845 --name aergo_test aergo/node
echo "Starting indexer"
sleep 3

AERGO_URL="localhost:7845"
ES_URL="http://localhost:9200"
CHAIN_PREFIX="chain_"
SYNC_FROM=0
SYNC_TO=30000000

./bin/indexer -A $AERGO_URL --dbtype elastic --dburl $ES_URL --prefix $CHAIN_PREFIX --from $SYNC_FROM --to $SYNC_TO --conflict 10
