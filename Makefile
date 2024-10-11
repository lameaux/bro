SRC_DIR := .
BUILD_DIR := ./bin
GIT_HASH := $(shell git rev-parse --short HEAD)
DOCKER_REPO := ghcr.io
DOCKER_IMAGE := lameaux/bro
TEST_FLAGS := -race -coverprofile=coverage.out

GO_FILES := $(shell find $(SRC_DIR) -name '*.go' ! -path '$(SRC_DIR)/protos/*go')

.PHONY: all
all: clean build lint test

.PHONY: generate
generate:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protos/metrics/metrics.proto

.PHONY: build
build:
	go build -ldflags "-X main.GitHash=$(GIT_HASH)" -o $(BUILD_DIR)/bro $(SRC_DIR)/cmd/client/*.go
	go build -ldflags "-X main.GitHash=$(GIT_HASH)" -o $(BUILD_DIR)/brod $(SRC_DIR)/cmd/server/*.go

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
	mv $(BUILD_DIR)/bro $(GOPATH)/bin
	mv $(BUILD_DIR)/brod $(GOPATH)/bin

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
	docker build --build-arg GIT_HASH=$(GIT_HASH) -t $(DOCKER_IMAGE):$(GIT_HASH) .

.PHONY: docker-push
docker-push:
	docker tag $(DOCKER_IMAGE):$(GIT_HASH) ghcr.io/$(DOCKER_IMAGE):latest
	docker push ghcr.io/$(DOCKER_IMAGE):latest

.PHONY: docker-release
docker-release: docker-build docker-push

