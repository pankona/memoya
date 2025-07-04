.PHONY: build install lint test fmt generate clean build-server

# Build MCP client
build:
	go build -o memoya ./cmd/memoya

# Build Cloud Run server
build-server:
	go build -o memoya-server ./cmd/memoya-server

# Install MCP client
install:
	go install ./cmd/memoya

# Code generation from OpenAPI
generate:
	@echo "Installing oapi-codegen v2..."
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@echo "Creating output directories..."
	mkdir -p internal/generated/server internal/generated/client
	@echo "Generating server code from OpenAPI spec..."
	oapi-codegen -config api/server-config.yaml api/openapi.yaml
	@echo "Generating client code from OpenAPI spec..."
	oapi-codegen -config api/client-config.yaml api/openapi.yaml
	@echo "Generated code created successfully"

lint:
	go vet ./...

test:
	go test ./...

fmt:
	goimports -w .

clean:
	rm -f memoya memoya-server
	rm -rf generated/

# Docker build for Cloud Run
docker-build:
	docker build -t memoya-server .

# Run server locally
run-server:
	go run ./cmd/memoya-server

# Run MCP client locally
run-client:
	go run ./cmd/memoya