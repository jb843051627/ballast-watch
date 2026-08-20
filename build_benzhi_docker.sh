#!/usr/bin/env bash
# 构建 ballast-watch Docker 镜像
# 用法: ./build_benzhi_docker.sh <NAME> <PLATFORM>
#   NAME    镜像名后缀，如 ballast-watch-bug-1
#   PLATFORM linux/amd64 或 linux/arm64
set -euo pipefail

NAME="${1:?用法: build_benzhi_docker.sh <NAME> <PLATFORM>}"
PLATFORM="${2:-linux/amd64}"

IMAGE="benzhi/${NAME}:latest"

echo "==> 构建 ${IMAGE} (${PLATFORM})"
docker build --platform "${PLATFORM}" -t "${IMAGE}" -f benzhi.Dockerfile .
echo "==> 完成: ${IMAGE}"
docker images | grep "benzhi/${NAME}" || true