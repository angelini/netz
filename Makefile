.PHONY: setup build
.PHONY: generate build-debug-server build-all-images
.PHONY: curl-front-proxy
.PHONY: k3d-create
.PHONY: k8s-clean k8s-deploy

CLUSTER := netz
INGRESS_PORT := 8001

REGISTRY := netzregistry
REIGSTRY_PORT := 3001

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

build-all-images: build-debug-server
	scripts/build-all.sh

curl-front-proxy:
	curl -i --connect-to api.example.com:80:localhost:$(INGRESS_PORT) api.example.com:80

k3d-create:
	k3d registry create $(REGISTRY).localhost --port $(REIGSTRY_PORT)
	k3d cluster create $(CLUSTER) -p "$(INGRESS_PORT):80@loadbalancer" --registry-use k3d-$(REGISTRY).localhost:$(REIGSTRY_PORT)

k8s-clean:
	kubectl delete ing --ignore-not-found ingress
	kubectl delete deployment --ignore-not-found api
	kubectl delete deployment --ignore-not-found front-proxy
	kubectl delete service --ignore-not-found api-svc
	kubectl delete service --ignore-not-found front-proxy-svc

k8s-deploy:
	kubectl apply -f k8s/api.yaml
	kubectl apply -f k8s/front-proxy.yaml