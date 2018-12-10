all: build/esindexer

protoc:
	protoc -I./aergo-protobuf/proto --go_out=plugins=grpc,paths=source_relative:./types ./aergo-protobuf/proto/*.proto

build/esindexer: *.go protoc
	go build -o build/esindexer main.go

run:
	go run main.go

run-reindex:
	go run main.go --reindex