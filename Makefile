.PHONY: all
all: vet lint test build

.PHONY: build
build:
	go build ./cmd/sfn-depends

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -rf sfn-depends
