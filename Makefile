DUCKDB_VERSION=0.4.0
LIB_PATH := $(shell pwd)/lib
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
	mkdir -p lib
	curl -Lo lib/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-$(ARCH_OS).zip
	cd lib; unzip -u libduckdb.zip

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
	$(LDFLAGS) go build -o 4wings -ldflags="-r $(LIB_PATH)" main.go

.PHONY: release-mac
release-mac: $(LIBS)
	make ./dist
	$(LDFLAGS) GOOS=darwin GOARCH=arm64 go build  -o ./dist/${NAME}-arm64 -ldflags="-r $(LIB_PATH)" main.go
	$(LDFLAGS) GOOS=darwin GOARCH=amd64 go build  -o ./dist/${NAME}-amd64 -ldflags="-r $(LIB_PATH)" main.go
	lipo -create -output ./dist/${NAME}-osx ./dist/${NAME}-amd64 ./dist/${NAME}-arm64

.PHONY: clean
clean:
	rm -rf lib
	rm -rf dist