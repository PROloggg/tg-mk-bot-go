.PHONY: build run clean prepare

BINARY_NAME=course-bot
BUILD_DIR=build

prepare:
	mkdir -p $(BUILD_DIR)/db
	cp .env $(BUILD_DIR)/
	cp -r data $(BUILD_DIR)/

build: prepare export-db
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./

export-db: prepare
	go build -o $(BUILD_DIR)/$(EXPORT_DB_BINARY) ./cmd/exportDb.go

run: build
	cd $(BUILD_DIR) && ./$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)

build-win: prepare export-db-win
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe ./

export-db-win: prepare
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(EXPORT_DB_BINARY).exe ./cmd/exportDb.go
