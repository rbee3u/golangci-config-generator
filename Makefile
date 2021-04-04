lint:
	go run ./cmd/golangci-config-generator
	golangci-lint run

install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1
