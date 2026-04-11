.PHONY: help dev build test lint docker clean

help:
	@echo "SearchJet OSS - Make commands"
	@echo ""
	@echo "  dev      Start local development (requires Docker)"
	@echo "  build    Build the server binary"
	@echo "  test     Run unit tests"
	@echo "  lint     Run golangci-lint"
	@echo "  docker   Build Docker image"
	@echo "  clean    Clean build artifacts"

dev:
	docker compose -f deploy/docker-compose.yml up -d

build:
	go build -o bin/searchjet ./cmd/server

test:
	go test -v ./...

lint:
	golangci-lint run ./...

docker:
	docker build -t searchjet-oss/searchjet:latest -f deploy/Dockerfile .

clean:
	rm -rf bin/