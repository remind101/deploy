.PHONY: cmd

cmd:
	go build -o build/deploy ./cmd/deploy

vet:
	go vet $(shell go list ./... | grep -v /vendor/)

test:
	go test -race $(shell go list ./... | grep -v /vendor/)

release:
	./scripts/release $(VERSION)
