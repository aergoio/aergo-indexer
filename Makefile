all: bin/esindexer

protoc:
	protoc -I./aergo-protobuf/proto --go_out=plugins=grpc,paths=source_relative:./types ./aergo-protobuf/proto/*.proto

bin/esindexer: *.go indexer/*.go indexer/**/*.go types/*.go
	go build -o bin/esindexer main.go

run:
	go run main.go

run-reindex:
	go run main.go --reindex