GO       ?= go
BINARY   := bin/dataflute
PACKAGES := ./cmd/... ./internal/... ./tests/...

.PHONY: all build run test test-race lint fmt clean install docker

all: fmt lint test build

build:
	$(GO) build -o $(BINARY) ./cmd/dataflute

run:
	$(GO) run ./cmd/dataflute serve --data ./pkg/data --host 0.0.0.0 --port 8080

test:
	$(GO) test -cover $(PACKAGES)

# Race detection requires cgo; not enabled on all toolchains.
test-race:
	$(GO) test -race -cover $(PACKAGES)

lint:
	$(GO) vet $(PACKAGES)

fmt:
	gofmt -l -w .

clean:
	rm -rf bin

install:
	$(GO) install ./cmd/dataflute

docker:
	docker build -t dataflute .
