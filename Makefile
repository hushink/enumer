.PHONY: build test race lint check golden tidy license vuln

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

license:
	go run github.com/apache/skywalking-eyes/cmd/license-eye@v0.9.0 header check
	go run github.com/apache/skywalking-eyes/cmd/license-eye@v0.9.0 dependency check

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
