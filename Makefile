DUCKDB_VERSION=0.4.0
LIB_PATH := /var/lib/libduckdb
NAME=geotemporal-data-explorer

ifeq ($(shell uname -s),Darwin)
LIB_EXT=dylib
ARCH_OS=osx-universal
LIBRARY_PATH := DYLD_LIBRARY_PATH=$(LIB_PATH)
else
LIB_EXT=so
ARCH_OS=linux-amd64
LIBRARY_PATH := LD_LIBRARY_PATH=$(LIB_PATH)
endif
LIBS := lib/libduckdb.$(LIB_EXT)
LDFLAGS := LIB=libduckdb.$(LIB_EXT) CGO_LDFLAGS="-L$(LIB_PATH)" $(LIBRARY_PATH) CGO_CFLAGS="-I$(LIB_PATH)"

$(LIBS):
	mkdir -p /var/lib/libduckdb
	curl -Lo /var/lib/libduckdb/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-$(ARCH_OS).zip
	unzip -u /var/lib/libduckdb/libduckdb.zip -d /var/lib/libduckdb

mac-libs:
	mkdir -p /var/lib/libduckdb
	curl -Lo /var/lib/libduckdb/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-osx-universal.zip
	unzip -u /var/lib/libduckdb/libduckdb.zip -d /var/lib/libduckdb

linux-libs:
	mkdir -p /var/lib/libduckdb
	curl -Lo /var/lib/libduckdb/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-linux-amd64.zip
	unzip -u /var/lib/libduckdb/libduckdb.zip -d /var/lib/libduckdb

ui:
	mkdir -p ./dist
	find ./assets -type f ! -name "*.go" -exec rm {} \;
	rm -rf ./assets/_next
	rm -rf ./assets/icons

	curl -Lo ./assets/ui.zip https://storage.googleapis.com/geotemporal-data-explorer-releases/ui/latest.zip
	unzip -u ./assets/ui.zip -d ./assets
	rm -rf ./assets/ui.zip
	rm -rf ./assets/__MACOSX
.PHONY: install
install: $(LIBS)
	$(LDFLAGS) go install -ldflags="-r $(LIB_PATH)" ./...

.PHONY: develop
develop: $(LIBS)
	$(LDFLAGS) ENV=dev reflex -r "\.go$"" -s  -- go run -ldflags="-r $(LIB_PATH)" main.go -v server  

.PHONY: test
test: $(LIBS)
	$(LDFLAGS) go test -ldflags="-r $(LIB_PATH)" -v -race -count=1 ./...

.PHONY: build
build: $(LIBS)
	$(LDFLAGS) go build -o geotemporal-data-explorer -ldflags="-r $(LIB_PATH)" main.go

.PHONY: release-linux
release-linux: ui linux-libs
	mkdir -p ./dist
	$(LDFLAGS) GOOS=linux GOARCH=amd64 go build  -o ./dist/${NAME}-linux-amd64-${VERSION} -ldflags="-r $(LIB_PATH)" main.go
	
.PHONY: release-mac
release-mac: ui mac-libs
	$(LDFLAGS) GOOS=darwin GOARCH=arm64 go build  -o ./dist/${NAME}-darwin-arm64-${VERSION} -ldflags="-r $(LIB_PATH)" main.go
	

.PHONY: clean
clean:
	rm -rf lib
	rm -rf dist