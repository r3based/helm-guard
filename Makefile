BINARY=helm-guard
CMD_PATH=./cmd/helm-guard
BIN_DIR=bin

.PHONY: build run test lint fmt clean install

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD_PATH)

run:
	go run $(CMD_PATH)

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

install:
	go install $(CMD_PATH)

clean:
	rm -rf $(BIN_DIR)