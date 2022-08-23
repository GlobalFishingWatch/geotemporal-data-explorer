echo "Downloading UI"
find ./assets -type f ! -name "*.go" -exec rm {} \;
rm -rf ./assets/_next
rm -rf ./assets/icons

curl -Lo ./assets/ui.zip https://storage.googleapis.com/geotemporal-data-explorer-releases/ui/latest.zip
unzip -u ./assets/ui.zip -d ./assets
rm -rf ./assets/ui.zip

echo "Creating release in mac $1"
make VERSION=$1 release-mac 

gsutil cp ./dist/geotemporal-data-explorer-darwin-arm64-$1 gs://geotemporal-data-explorer-releases/versions/geotemporal-data-explorer-darwin-arm64-$1
gsutil cp ./dist/geotemporal-data-explorer-darwin-arm64-$1 gs://geotemporal-data-explorer-releases/versions/geotemporal-data-explorer-darwin-arm64-latest

echo "Creating tag in git"
git tag $1