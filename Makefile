.PHONY: build test lint clean

build:
	go build ./...

test:
	go test ./... -count=1 -timeout 30s

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
