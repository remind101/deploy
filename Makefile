.PHONY: cmd

cmd:
	godep go build -o build/deploy ./cmd/deploy

test:
	godep go test -race ./...

release:
	./scripts/release $(VERSION)
