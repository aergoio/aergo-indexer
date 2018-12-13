FROM golang:alpine as builder
RUN apk update && apk add git glide build-base
ENV GOPATH $HOME/go
RUN go get -d github.com/aergoio/aergo-esindexer && cd ${GOPATH}/src/github.com/aergoio/aergo-esindexer && git submodule init && git submodule update && make all

FROM alpine:3.8
RUN apk add libgcc
COPY --from=builder $HOME/go/src/github.com/aergoio/aergo-esindexer/bin/* /usr/local/bin/
CMD ["esindexer"]