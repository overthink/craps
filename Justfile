lint:
  golangci-lint run ./...

build:
  mkdir -p bin
  go build -o bin/craps ./cmd

test:
  go test ./...
