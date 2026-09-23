.PHONY: build clean test install

VERSION ?= 1.0.0
BIN_DIR = bin
TARGET = $(BIN_DIR)/deepmd

build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(TARGET) ./cmd/deepmd
	@echo "Built $(TARGET) version $(VERSION)"

test:
	go test -v ./...

install: build
	@mkdir -p $(HOME)/.local/bin
	cp -f $(TARGET) $(HOME)/.local/bin/deepmd
	@chmod +x $(HOME)/.local/bin/deepmd
	@echo "Installed deepmd to $(HOME)/.local/bin/deepmd"

clean:
	rm -rf $(BIN_DIR)
