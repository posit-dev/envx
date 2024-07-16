IMAGE_REGISTRY := 'ghcr.io/posit-dev'
IMAGE_NAME := 'envx'
BUILDX_PATH := ''
VERSION := `git describe --always --dirty --tags`

default: build test

test *ARGS:
  go test -v -coverprofile cover.out {{ARGS}} ./...


build *ARGS:
  mkdir -p ./build && \
  CGO_ENABLED=0 go build -tags netgo -a -o build/envx {{ARGS}} ./cmd/envx/main.go

smoke *ARGS:
  build/envx --help

docker-build:
  #!/usr/bin/env bash
  set -o errexit
  set -o pipefail
  BUILDER=''
  declare -a BUILDX_ARGS
  if [[ -n '{{ BUILDX_PATH }}' ]]; then
    BUILDER='--builder={{ BUILDX_PATH }}'
    BUILDX_ARGS=(
      '--cache-from=type=local,src=/tmp/.buildx-cache'
      '--cache-to=type=local,dest=/tmp/.buildx-cache'
    )
  fi
  set -o xtrace
  docker buildx ${BUILDER} build --load "${BUILDX_ARGS[@]}" \
    --platform='linux/amd64' \
    -t '{{ IMAGE_REGISTRY }}/{{ IMAGE_NAME }}:{{ VERSION }}' \
    .

docker-push:
  docker push {{ IMAGE_REGISTRY }}/{{ IMAGE_NAME }}:{{ VERSION }}

echo-image:
  @echo {{ IMAGE_NAME }}:{{ VERSION }}
