FROM golang:1.24.0-alpine AS builder
WORKDIR /workspace
COPY go.mod .
COPY go.sum .
RUN apk add --no-cache alpine-sdk libsodium-dev zeromq-dev czmq-dev && go mod download
COPY main.go .
RUN go build

FROM alpine:latest
RUN apk add --no-cache libstdc++ libgcc libsodium libzmq czmq
COPY --from=builder /workspace/magnetico-bitmagnet /magnetico-bitmagnet
ENTRYPOINT ["/magnetico-bitmagnet"]