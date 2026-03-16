# Makefile
.PHONY: build run clean dev

build:
	@echo "Building binary..."
	@go build -o bin/adapter_sidecar cmd/service/main.go

run: build
	@echo "Running..."
	@./bin/artifacts-svc

clean:
	@rm -rf bin/

dev:
	@echo "Starting dev server..."
	@/home/sirkartik/go/bin/air
