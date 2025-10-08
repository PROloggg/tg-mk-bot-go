.PHONY: build run clean prepare

BINARY_NAME=course-bot
BUILD_DIR=build
EXPORT_DB_BINARY=exportDB

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
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe ./

export-db-win: prepare
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(EXPORT_DB_BINARY).exe ./cmd/exportDb.go
CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 