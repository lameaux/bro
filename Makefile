APP_NAME := bro
SRC_DIR := .
BUILD_DIR := ./bin
BUILD_FILE := $(BUILD_DIR)/$(APP_NAME)
GIT_HASH := $(shell git rev-parse --short HEAD)

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -ldflags "-X main.GitHash=$(GIT_HASH)" -o $(BUILD_FILE) $(SRC_DIR)/**/*.go

.PHONY: install
install: build
	mv $(BUILD_FILE) $(GOPATH)/bin

run: build
	$(BUILD_FILE) $(ARGS)

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)