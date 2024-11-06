SRC_DIR := .
BUILD_DIR := ./bin

VERSION := v0.1.0
BUILD_HASH := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

DOCKER_REPO := ghcr.io
DOCKER_IMAGE := lameaux/bro
DOCKER_BUILD := docker buildx build --platform linux/amd64,linux/arm64
DOCKER_TAG := latest

TEST_FLAGS := -race -coverprofile=coverage.out

GO_FILES := $(shell find $(SRC_DIR) -name '*.go' ! -path '$(SRC_DIR)/protos/*go')

.PHONY: all
all: build lint test

.PHONY: generate
generate:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protos/metrics/metrics.proto

.PHONY: build
build: clean
	go build -ldflags "-X main.Version=$(VERSION) -X main.BuildHash=$(BUILD_HASH) -X main.BuildDate=$(BUILD_DATE)" \
		-o $(BUILD_DIR)/bro $(SRC_DIR)/cmd/client/*.go
	go build -ldflags "-X main.Version=$(VERSION) -X main.BuildHash=$(BUILD_HASH) -X main.BuildDate=$(BUILD_DATE)" \
		-o $(BUILD_DIR)/brod $(SRC_DIR)/cmd/server/*.go

.PHONY: fmt
fmt:
	gci write $(GO_FILES) --skip-generated -s standard -s default
	gofumpt -l -w $(GO_FILES)

.PHONY: fmt-install
fmt-install:
	go install github.com/daixiang0/gci@latest
	go install mvdan.cc/gofumpt@latest

.PHONY: lint
lint:
	golangci-lint run

.PHONY: lint-install
lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: test
test:
	go test $(SRC_DIR)/... $(TEST_FLAGS)

.PHONY: coverage
coverage:
	go tool cover -func coverage.out | grep "total:" | \
	awk '{print ((int($$3) > 80) != 1) }'


.PHONY: install
install: build
	cp $(BUILD_DIR)/bro $(GOPATH)/bin
	cp $(BUILD_DIR)/brod $(GOPATH)/bin

.PHONY: run
run: build
	$(BUILD_DIR)/bro $(ARGS)

.PHONY: serve
serve: build
	$(BUILD_DIR)/brod $(ARGS)

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: docker-build
docker-build:
	$(DOCKER_BUILD) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_HASH=$(BUILD_HASH) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE):$(VERSION)-$(BUILD_HASH) .

.PHONY: docker-push
docker-push:
	docker tag $(DOCKER_IMAGE):$(VERSION)-$(BUILD_HASH) $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-release
docker-release: docker-build docker-push

.PHONY: docker-run
docker-run:
	docker run --rm $(DOCKER_REPO)/$(DOCKER_IMAGE):$(DOCKER_TAG) $(ARGS)