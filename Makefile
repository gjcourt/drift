.PHONY: build run dev test fmt vet lint check tidy clean

BINARY := drift
CMD    := ./cmd/drift

## build: compile the binary
build:
	go build -o $(BINARY) $(CMD)

## run: build then run
run: build
	./$(BINARY)

## dev: run with hot-reload friendly flags
dev:
	DRIFT_ADDR=:8080 go run $(CMD)

## test: run all unit tests
test:
	go test -race -count=1 ./...

## fmt: format all Go source files with gofmt
fmt:
	go fmt ./...

## vet: run go vet across all packages
vet:
	go vet ./...

## lint: run golangci-lint (install: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

## check: fmt + vet + lint + test
check: fmt vet lint test

## tidy: tidy and verify go.mod / go.sum
tidy:
	go mod tidy
	go mod verify

## clean: remove build artefacts
clean:
	rm -f $(BINARY) drift.db
