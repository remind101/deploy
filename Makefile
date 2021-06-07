.DEFAULT_GOAL := build/deploy

build/deploy: *.go cmd/deploy/*.go vet test
	go build -o build/deploy ./cmd/deploy

vet:
	go vet $(shell go list ./... | grep -v /vendor/)

test:
	go test -race $(shell go list ./... | grep -v /vendor/)

.PHONY: install
install: build/deploy
	install -T build/deploy /usr/local/bin/deploy

release:
	./scripts/release $(VERSION)
