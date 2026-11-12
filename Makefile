.PHONY: build test race lint check golden tidy

build:
	go build -o bin/enumer ./cmd/enumer

test:
	go test ./...

race:
	go test -race ./...

lint:
	golangci-lint run ./...

check:
	go run ./cmd/enumer check ./...

golden:
	go test ./generator -run TestGolden -update

tidy:
	go mod tidy
