.PHONY: help proto all
DEFAULT_GOAL: help

all: client server ## Build the client and server

client: ## Build the client
	go build ./cmd/crawler

server: ## Build the server
	go build ./cmd/crawler-service


proto: ## Generate the protobuf code
	mkdir -p pkg/crawler
	docker run -ti --rm \
		-u $(shell id -u):$(shell id -g) \
		-v $(shell pwd):/crawler \
		-w /crawler \
		grpc/go:latest \
		protoc \
			--proto_path ./api/protobuf/v1/crawler \
			--go_out=plugins=grpc:pkg/crawler \
			--go_opt=paths=source_relative \
			./api/protobuf/v1/crawler/crawler.proto

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
