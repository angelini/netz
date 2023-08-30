.PHONY: setup build

ENVOY_VERSION := 1.24.0

CMD_GO_FILES := $(shell find cmd/ -type f -name '*.go')
PKG_GO_FILES := $(shell find pkg/ -type f -name '*.go')

bin/envoy:
	@mkdir -p bin
	curl -fsSL -o bin/envoy https://github.com/envoyproxy/envoy/releases/download/v$(ENVOY_VERSION)/envoy-$(ENVOY_VERSION)-linux-x86_64
	chmod +x bin/envoy

setup: bin/envoy

bin/netz: $(CMD_GO_FILES) $(PKG_GO_FILES)
	@mkdir -p bin
	go mod tidy
	go build -o $@ main.go

bin/debug_server: debug_server/main.go
	@mkdir -p bin
	go build -o $@ $<

build: bin/netz bin/debug_server

generate:
	go run main.go generate all