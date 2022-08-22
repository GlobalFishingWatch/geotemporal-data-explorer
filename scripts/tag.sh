echo "Creating release in mac $1"
make VERSION=$1 release-mac 

echo "Creating tag in git"
git tag $1