.PHONY: all test build clean run dev

all: run

clean:
	@rm -rf build/ && go clean -testcache

build:
	@mkdir -p build && go build -o build/main .

test:
	@go test ./... -v -cover

run: build
	./build/main

dev:
	@air -c .air.toml || echo "Install air: go install github.com/cosmtrek/air@latest"

