#!/bin/bash
tool=midi2key-ng
version=1.1.0
dir=build

mkdir build

# Windows amd64
goos=windows
goarch=amd64
GOOS=$goos GOARCH=$goarch go build -o midi2key-ng.exe
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip midi2key-ng.exe

# Linux amd64
goos=linux
goarch=amd64
GOOS=$goos GOARCH=$goarch go build -o midi2key-ng
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip midi2key-ng

# Linux arm64
goos=linux
goarch=arm64
GOOS=$goos GOARCH=$goarch go build -o midi2key-ng
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip midi2key-ng

# Darwin/MacOS amd64
goos=darwin
goarch=amd64
GOOS=$goos GOARCH=$goarch go build -o midi2key-ng
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip midi2key-ng

# Darwin/MacOS arm64
goos=darwin
goarch=arm64
GOOS=$goos GOARCH=$goarch go build -o midi2key-ng
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip midi2key-ng

# remove wcvs
rm midi2key-ng
rm midi2key-ng.exe

# generate checksum file
find "$dir"/ -type f  \( -iname "*.zip" \) -exec sha256sum {} + > "$dir"/"$tool"_"$version"_checksums_sha256.txt