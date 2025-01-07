SHELL := bash

REPOSITORY ?= localhost
CONTAINER_NAME ?= minecraft-exporter
TAG ?= latest

build:
	hack/build.sh

image:
	podman build -t $(REPOSITORY)/$(CONTAINER_NAME):$(TAG) .

test:
	go test -v -coverprofile=coverprofile.out ./...

coverprofile:
	hack/coverprofile.sh

lint:
	golangci-lint run -v

fmt:
	gofmt -s -w ./cmd ./pkg

validate:
	hack/validate.sh

update-deps:
	hack/update-deps.sh

clean:
	rm -rf bin coverprofiles coverprofile.out

.PHONY: \
	default \
	build \
	image \
	test \
    coverprofile \
	lint \
	fmt \
	validate \
	update-deps \
    clean \
	$(NULL)
