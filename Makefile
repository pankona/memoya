.PHONY: lint test fmt

lint:
	go vet ./...

test:
	go test ./...

fmt:
	goimports -w .