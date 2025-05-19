lint:
  golangci-lint run ./...

build:
  mkdir -p bin
  go build -o bin/craps ./cmd

test:
  go test ./...

# stats assumes only a single strategy being tested at once
stats:
  go run ./cmd --trials=100000 | tail -n +2 | cut -d, -f2 | sta -q --transpose --fixed

[script]
codex:
  OPENAI_API_KEY=$(<~/.codex-api-key) codex $@
