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

lint:
	golangci-lint run -v

fmt:
	gofmt -s -w ./cmd ./pkg

validate:
	hack/validate.sh

update-deps:
	hack/update-deps.sh

.PHONY: \
	default \
	build \
	build-image \
	test \
	lint \
	fmt \
	validate \
	update-deps \
	$(NULL)
