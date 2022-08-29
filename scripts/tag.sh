echo "Creating release in mac $1"
sudo make VERSION=$1 release-mac 

gsutil cp ./dist/geotemporal-data-explorer-darwin-arm64-$1 gs://geotemporal-data-explorer-releases/versions/geotemporal-data-explorer-darwin-arm64-$1
gsutil cp ./dist/geotemporal-data-explorer-darwin-arm64-$1 gs://geotemporal-data-explorer-releases/versions/geotemporal-data-explorer-darwin-arm64-latest

echo "Creating tag ($1) in git"
git tag $1
git push origin $1