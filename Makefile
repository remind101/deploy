.PHONY: cmd

help: #! Show this help message.
	@echo 'Usage: make [OPTIONS] [TARGET]'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/\(^.*:\).*#!\( .*\$\)/  \1\2/p' $(MAKEFILE_LIST) | column -t -s ':'

cmd: #! Build executable
	go build -o build/deploy ./cmd/deploy

vet: #! Run linters and style checkers
	go vet ./...

test: #! Run all tests
	go test -race ./...

release: # Package releases
	./scripts/release $(VERSION)
