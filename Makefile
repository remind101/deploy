.PHONY: cmd

cmd:
	godep go build -o build/deploy ./cmd/deploy

release:
	goxc -bc="linux,darwin" -d build -pv="0.0.3"
