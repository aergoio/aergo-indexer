#!/bin/sh
VERSION=${VERSION:=6.4.0}

#docker rm -f es_test
#docker rm -f aergo_test

#echo "Starting elasticsearch"
docker run -d -p 9200:9200 --name es_test -e "http.host=0.0.0.0" -e "transport.host=127.0.0.1" -e "bootstrap.memory_lock=true" -e "ES_JAVA_OPTS=-Xms1g -Xmx1g" docker.elastic.co/elasticsearch/elasticsearch-oss:$VERSION elasticsearch -Enetwork.host=_local_,_site_ -Enetwork.publish_host=_local_

echo "Starting mariadb"
docker run -d -p 3306:3306 --name mariadb_test -e MYSQL_ROOT_PASSWORD=my-secret-pw mariadb:10.4-bionic

echo "Starting aergosvr"
docker run -d -p 7845:7845 --name aergo_test aergo/node:1.2 aergosvr --config /aergo/testmode.toml

#docker run -d -p 7845:7845 --name aergo_test aergo/node
echo "Starting esindexer"
sleep 3

##./bin/esindexer -A testnet.aergo.io:7845 --dbtype mariadb -D "root:my-secret-pw@tcp(localhost:3306)/test" --prefix chain_ --reindex --from 905000 --to 909000 --exit-on-complete
./bin/esindexer -A testnet.aergo.io:7845 --dbtype mariadb -D "root:my-secret-pw@tcp(localhost:3306)/test" --prefix chain_ --reindex --from 0 --to 4499 --exit-on-complete
#./bin/esindexer -A testnet.aergo.io:7845 --dbtype es -D "http://localhost:9200" --prefix chain_ --reindex --from 0 --to 4499 --exit-on-complete

#./bin/esindexer -H localhost -p 7845 --dbtype mariadb -D "root:my-secret-pw@tcp(localhost:3306)/test" --prefix chain_
#./bin/esindexer -H localhost -p 7845 -D http://localhost:9200 --prefix chain_
# ./bin/esindexer -H localhost -p 7845 -D http://localhost:9200 --prefix chain_ --reindex
#./bin/esindexer -H localhost -p 7845 -D http://localhost:9200 -A testnet.aergo.io:7845 --reindex --prefix chain_
# docker rm -f es_test
# docker rm -f aergo_test
# docker rm -f mariadb_test