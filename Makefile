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

.PHONY: install
install: $(LIBS)
	$(LDFLAGS) go install -ldflags="-r $(LIB_PATH)" ./...

.PHONY: develop
develop: $(LIBS)
	$(LDFLAGS) reflex -r "\.go$"" -s  -- go run -ldflags="-r $(LIB_PATH)" main.go -v server 

.PHONY: test
test: $(LIBS)
	$(LDFLAGS) go test -ldflags="-r $(LIB_PATH)" -v -race -count=1 ./...

.PHONY: build
build: $(LIBS)
	$(LDFLAGS) go build -o geotemporal-data-explorer -ldflags="-r $(LIB_PATH)" main.go

.PHONY: release-linux
release-linux: $(LIBS)
	mkdir -p ./dist
	$(LDFLAGS) GOOS=linux GOARCH=amd64 go build  -o ./dist/${NAME}-linux-amd64 -ldflags="-r $(LIB_PATH)" main.go
	
.PHONY: release-mac
release-mac: mac-libs
	mkdir -p ./dist
	$(LDFLAGS) GOOS=darwin GOARCH=arm64 go build  -o ./dist/${NAME}-darwin-arm64-${VERSION} -ldflags="-r $(LIB_PATH)" main.go
	

.PHONY: clean
clean:
	rm -rf lib
	rm -rf dist