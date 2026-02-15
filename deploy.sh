#!/bin/bash
echo "Starting build..."
docker buildx build -t ghcr.io/westhecool/magnetico-bitmagnet:amd64 --platform linux/amd64 . --load
docker buildx build -t ghcr.io/westhecool/magnetico-bitmagnet:arm64 --platform linux/arm64 . --load
docker buildx build -t ghcr.io/westhecool/magnetico-bitmagnet:armv7 --platform linux/arm/v7 . --load

echo "Copying binaries..."
docker container rm temp_container -f &>/dev/null
docker create --name temp_container ghcr.io/westhecool/magnetico-bitmagnet:amd64 &>/dev/null
docker cp temp_container:/magnetico-bitmagnet magnetico-bitmagnet-linux-amd64
docker container rm temp_container -f &>/dev/null
docker create --name temp_container ghcr.io/westhecool/magnetico-bitmagnet:arm64 &>/dev/null
docker cp temp_container:/magnetico-bitmagnet magnetico-bitmagnet-linux-arm64
docker container rm temp_container -f &>/dev/null
docker create --name temp_container ghcr.io/westhecool/magnetico-bitmagnet:armv7 &>/dev/null
docker cp temp_container:/magnetico-bitmagnet magnetico-bitmagnet-linux-armv7
docker container rm temp_container -f &>/dev/null

echo "Pushing images..."
docker buildx build -t ghcr.io/westhecool/magnetico-bitmagnet:amd64 --platform linux/amd64 . --push
docker buildx build -t ghcr.io/westhecool/magnetico-bitmagnet:arm64 --platform linux/arm64 . --push
docker buildx build -t ghcr.io/westhecool/magnetico-bitmagnet:armv7 --platform linux/arm/v7 . --push

echo "Pushing Universal image..."
docker buildx build -t ghcr.io/westhecool/magnetico-bitmagnet:latest --platform linux/amd64,linux/arm64,linux/arm/v7 . --push

echo "Done."