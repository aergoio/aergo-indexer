FROM golang:1.12-alpine3.9 as builder
RUN apk update && apk add git glide build-base
ENV GOPATH $HOME/go
ENV GO111MODULE=on
WORKDIR ${GOPATH}/src/github.com/aergoio/aergo-indexer

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . .
RUN make all

FROM alpine:3.9
RUN apk add libgcc
COPY --from=builder $HOME/go/src/github.com/aergoio/aergo-indexer/bin/* /usr/local/bin/
ADD arglog.toml $HOME
CMD ["indexer"]