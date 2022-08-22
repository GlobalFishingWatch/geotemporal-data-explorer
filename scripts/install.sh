SO=$(uname -s)
DUCKDB_VERSION=0.4.0

mkdir -p /var/lib/libduckdb

	unzip -u /var/lib/libduckdb/libduckdb.zip -d /var/lib/libduckdb
if [ "$SO" == "Darwin" ]; then
    curl -Lo /var/lib/libduckdb/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-osx-universal.zip
    wget -O /usr/local/bin/geotemporal-data-explorer https://storage.googleapis.com/geotemporal-data-explorer-releases/versions/geotemporal-data-explorer-darwin-arm64-latest
    chmod +x /usr/local/bin/geotemporal-data-explorer
else
    curl -Lo /var/lib/libduckdb/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v${DUCKDB_VERSION}/libduckdb-linux-amd64.zip
    wget -O /bin/geotemporal-data-explorer https://storage.googleapis.com/geotemporal-data-explorer-releases/versions/geotemporal-data-explorer-linux-amd64-latest
    chmod +x /bin/geotemporal-data-explorer
fi
unzip -u /var/lib/libduckdb/libduckdb.zip -d /var/lib/libduckdb
rm -rf /var/lib/libduckdb/libduckdb.zip

echo "Close and reopen your terminal to start using geotemporal-data-explorer or load your preferences again"
