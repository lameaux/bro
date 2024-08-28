APP_NAME := bro
SRC_DIR := .
BUILD_DIR := ./bin
BUILD_FILE := $(BUILD_DIR)/$(APP_NAME)
GIT_HASH := $(shell git rev-parse --short HEAD)
DOCKER_REPO := ghcr.io
DOCKER_IMAGE := lameaux/bro

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -ldflags "-X main.GitHash=$(GIT_HASH)" -o $(BUILD_FILE) $(SRC_DIR)/**/*.go

.PHONY: install
install: build
	mv $(BUILD_FILE) $(GOPATH)/bin

.PHONY: run
run: build
	$(BUILD_FILE) $(ARGS)

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

