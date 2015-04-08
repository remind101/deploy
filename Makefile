.PHONY: cmd

cmd:
	godep go build -o build/deploy ./cmd/deploy
