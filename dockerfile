FROM golang:alpine3.20 AS builder
ENV CGO_ENABLED=1
ENV CGO_CFLAGS=-D_LARGEFILE64_SOURCE
WORKDIR /workspace
COPY go.mod .
COPY go.sum .
RUN apk add --no-cache alpine-sdk libsodium-dev zeromq-dev czmq-dev && go mod download
COPY main.go .
RUN go build

FROM alpine:3.20
RUN apk add --no-cache libstdc++ libgcc libsodium libzmq czmq
COPY --from=builder /workspace/magnetico-bitmagnet /magnetico-bitmagnet
ENTRYPOINT ["/magnetico-bitmagnet"]