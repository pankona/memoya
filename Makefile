.PHONY: build install lint test fmt

build:
	go build -o memoya ./cmd/memoya

install:
	go install ./cmd/memoya

lint:
	go vet ./...

test:
	go test ./...

fmt:
	goimports -w .