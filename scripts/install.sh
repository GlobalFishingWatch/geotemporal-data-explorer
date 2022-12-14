geotemporal_has() {
  type "$1" > /dev/null 2>&1
}

SO=$(uname -s)
DUCKDB_VERSION=0.4.0

mkdir -p /var/lib/libduckdb

if ! geotemporal_has "curl"; then
    echo "Error: curl command not found"
    exit 1
fi
if [ "$EUID" -ne 0 ]
  then echo "Please run as root to install libraries in /var/lib/"
  exit
fi


if [ "$SO" == "Darwin" ]; then
    rm -rf /usr/local/bin/geotemporal-data-explorer
    curl -Lo /var/lib/libduckdb/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-osx-universal.zip
    curl https://storage.googleapis.com/geotemporal-data-explorer-releases/versions/geotemporal-data-explorer-darwin-arm64-latest?rand=$RANDOM --output /usr/local/bin/geotemporal-data-explorer
    chmod +x /usr/local/bin/geotemporal-data-explorer
else
    rm -rf /bin/geotemporal-data-explorer
    curl -Lo /var/lib/libduckdb/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-linux-amd64.zip
    curl https://storage.googleapis.com/geotemporal-data-explorer-releases/versions/geotemporal-data-explorer-linux-amd64-latest?rand=$RANDOM --output /bin/geotemporal-data-explorer
    chmod +x /bin/geotemporal-data-explorer
fi
if geotemporal_has "unzip"; then
    unzip -u /var/lib/libduckdb/libduckdb.zip -d /var/lib/libduckdb 
    rm -rf /var/lib/libduckdb/libduckdb.zip
else
    echo "Error: unzip command not found"
    exit 1
fi

echo "Close and reopen your terminal to start using geotemporal-data-explorer or load your
preferences again"