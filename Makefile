all: bin/esindexer deps

protoc:
	protoc -I./aergo-protobuf/proto --go_out=plugins=grpc,paths=source_relative:./types ./aergo-protobuf/proto/*.proto

deps:
	glide install

bin/esindexer: *.go esindexer/*go types/*.go
	go build -o bin/esindexer main.go

run:
	go run main.go

run-reindex:
	go run main.go --reindex