.PHONY: all
all: bespoke lint tests

bespoke:
	go build -o bespoke ./cmd

tests:
	go test -race ./...

lint:
	golangci-lint run --timeout 2m
