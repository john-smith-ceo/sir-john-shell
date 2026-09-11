.PHONY: build build-all vet fmt clean release

APP = sir-john-shell
CMD = ./cmd/$(APP)
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -o $(APP) $(CMD)

linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/$(APP)-linux-amd64 $(CMD)

darwin-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o dist/$(APP)-darwin-arm64 $(CMD)

darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/$(APP)-darwin-amd64 $(CMD)

build-all: clean linux-amd64 darwin-arm64 darwin-amd64

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf dist/

release: build-all
	@echo "Binaries in dist/"
