.PHONY: cmd

cmd:
	go build -o build/deploy ./cmd/deploy

test:
	go test -race $(shell go list ./... | grep -v /vendor/)

release:
	./scripts/release $(VERSION)
