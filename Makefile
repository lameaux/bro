APP_NAME := loadpro
SRC_DIR := .
BUILD_DIR := ./bin
BUILD_FILE := $(BUILD_DIR)/$(APP_NAME)

.PHONY: all
all: build

.PHONY: build
build:
	go build -o $(BUILD_FILE) $(SRC_DIR)/**/*.go

run: build
	$(BUILD_FILE)

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)