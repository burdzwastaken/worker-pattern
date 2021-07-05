SHELL := /usr/bin/env bash -o errexit -o pipefail -o nounset

NAME     := worker-pattern
VERSION  := v0.0.1
DATE     := $(shell date -u)
CGO_FLAG := GCO_ENABLED=0
LDFLAGS  := -ldflags="-s -w -X 'main.version=$(VERSION)' -X 'main.buildDate=$(DATE)'"
export

.PHONY: default
default: lint test build up

.PHONY: lint
lint: lint-go

.PHONY: lint-docker
lint-docker:
	hadolint Dockerfile

.PHONY: lint-go
lint-go:
	go vet ./...
	go mod tidy

.PHONY: clean
clean:
	go clean -i -x
	rm -rf redis-data/

.PHONY: test
test: lint-go
	go test -v ./...

.PHONY: build
build: clean
	$(GCO_FLAG) go build $(LDFLAGS)

.PHONY: install
install: clean
	$(GCO_FLAG) go install $(LDFLAGS)

.PHONY: up
up:
	docker-compose up --build
