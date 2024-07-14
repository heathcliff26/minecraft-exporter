SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= minecraft-exporter
TAG ?= latest

default: build

build:
	hack/build.sh

build-image:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(TAG) .

test:
	go test -v ./...

update-deps:
	hack/update-deps.sh

.PHONY: \
	default \
	build \
	build-image \
	test \
	update-deps \
	$(NULL)
