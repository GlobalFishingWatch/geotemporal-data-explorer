echo "Creating release in mac $1"
make VERSION=$1 release-mac 

gsutil cp ./dist/geotemporal-data-explorer-darwin-arm64-$1 gs://raul-import/versions/$1
gsutil cp ./dist/geotemporal-data-explorer-darwin-arm64-$1 gs://raul-import/versions/$1/geotemporal-data-explorer-darwin-arm64-latest

echo "Creating tag in git"
git tag $1