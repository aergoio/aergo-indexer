#!/bin/sh
VERSION=${VERSION:=6.4.0}

#docker rm -f es_test
#docker rm -f aergo_test

echo "Starting elasticsearch"
docker run -d -p 9200:9200 --name es_test -e "http.host=0.0.0.0" -e "transport.host=127.0.0.1" -e "bootstrap.memory_lock=true" -e "ES_JAVA_OPTS=-Xms1g -Xmx1g" docker.elastic.co/elasticsearch/elasticsearch-oss:$VERSION elasticsearch -Enetwork.host=_local_,_site_ -Enetwork.publish_host=_local_
echo "Starting aergosvr"
docker run -d -p 7845:7845 --name aergo_test aergo/node:1.0.0-rc aergosvr --config /aergo/testmode.toml
echo "Starting esindexer"
sleep 3
./bin/esindexer -H localhost -p 7845 --esurl http://localhost:9200 --prefix chain_ --reindex
#./bin/esindexer -H localhost -p 7845 -E http://localhost:9200 -A testnet.aergo.io:7845 --reindex --prefix chain_
# docker rm -f es_test
# docker rm -f aergo_test