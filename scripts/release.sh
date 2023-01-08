VERSION=`git describe --tags`
PREFIX=sqlite-rest_${VERSION}_

set -e

echo "Building assets for the release ${VERSION}..."

# Cleanup relaese dir
rm -rf ./release

# Create temp dirs
mkdir ./release/ \
./release/${PREFIX}windows-amd64/ \
./release/${PREFIX}linux-amd64/ \
./release/${PREFIX}linux-arm/ \
./release/${PREFIX}linux-arm64/

# Copy assets
cp -R ./LICENSE ./README.md ./CHANGELOG.md ./release/${PREFIX}windows-amd64/
cp -R ./LICENSE ./README.md ./CHANGELOG.md ./release/${PREFIX}linux-amd64/
cp -R ./LICENSE ./README.md ./CHANGELOG.md ./release/${PREFIX}linux-arm/
cp -R ./LICENSE ./README.md ./CHANGELOG.md ./release/${PREFIX}linux-arm64/

# Build for each platform
GOOS=windows GOARCH=amd64 go build -o ./release/${PREFIX}windows-amd64/sqlite-rest.exe ./cmd/sqlite-rest.go &
GOOS=linux GOARCH=amd64 go build -o ./release/${PREFIX}linux-amd64/sqlite-rest ./cmd/sqlite-rest.go &
GOOS=linux GOARCH=arm64 go build -o ./release/${PREFIX}linux-arm64/sqlite-rest ./cmd/sqlite-rest.go &
GOOS=linux GOARCH=arm go build -o ./release/${PREFIX}linux-arm/sqlite-rest ./cmd/sqlite-rest.go &
wait

# Archive release folders
cd ./release/
zip -r ./${PREFIX}windows-amd64.zip ./${PREFIX}windows-amd64/ &
tar -czvf ./${PREFIX}linux-amd64.tar.gz ./${PREFIX}linux-amd64/ &
tar -czvf ./${PREFIX}linux-arm64.tar.gz ./${PREFIX}linux-arm64/ &
tar -czvf ./${PREFIX}linux-arm.tar.gz ./${PREFIX}linux-arm/ &
wait

# Destroy temp dirs
rm -rf ./${PREFIX}windows-amd64 &
rm -rf ./${PREFIX}linux-amd64 &
rm -rf ./${PREFIX}linux-arm64 &
rm -rf ./${PREFIX}linux-arm &
wait