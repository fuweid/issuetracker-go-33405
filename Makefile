NAME := Wei Fu
EMAIL := fuweid89@gmail.com

.PHONY: build
build: ## build the binary into bin folder
	go build -o bin/parent ./cmd/parent
	go build -o bin/child ./cmd/child

.PHONY: clean
clean: ## clean tmp and bin folder
	rm -rf ./tmp/*
	rm -rf ./bin/*

.PHONY: test
test: ## reproduce the issue
	@mkdir -p $${PWD}/tmp
	@bin/parent -root $${PWD}/tmp/ -childCmd bin/child

.PHONY: help
help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
