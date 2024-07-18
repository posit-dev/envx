IMAGE_REGISTRY ?= ghcr.io/posit-dev
IMAGE_NAME ?= envx
BUILDX_PATH ?=
VERSION := $(shell git describe --always --dirty --tags)

BUILD_OUTPUT := ./dist/envx

export IMAGE_REGISTRY
export IMAGE_NAME
export VERSION

.PHONY: all
all: build test

.PHONY: test
test:
	go test -v -coverprofile cover.out $(GO_TEST_ARGS) ./...

.PHONY: build
build: $(BUILD_OUTPUT)

$(BUILD_OUTPUT):
	mkdir -p ./dist && \
	CGO_ENABLED=0 go build -tags netgo -a -o $@ $(GO_BUILD_ARGS) ./cmd/envx/...

.PHONY: clean
clean:
	$(RM) -r ./dist

.PHONY: smoke
smoke: $(BUILD_OUTPUT)
	$(BUILD_OUTPUT) --help

.PHONY: docker-build
docker-build:
	./scripts/docker-build

.PHONY: docker-push
docker-push:
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(VERSION)

.PHONY: echo-image
echo-image:
	@echo $(IMAGE_NAME):$(VERSION)
