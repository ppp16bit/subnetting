BINARY := subnetting
CMD := ./cmd/subnetting
BIN_DIR := bin
GO ?= go
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: build run test vet fmt clean install uninstall

build:
	mkdir -p "$(BIN_DIR)"
	$(GO) build -o "$(BIN_DIR)/$(BINARY)" $(CMD)

run:
	$(GO) run $(CMD)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf -- "$(BIN_DIR)"

install:
	mkdir -p "$(INSTALL_DIR)"
	$(GO) build -o "$(INSTALL_DIR)/$(BINARY)" $(CMD)

uninstall:
	rm -f -- "$(INSTALL_DIR)/$(BINARY)"
