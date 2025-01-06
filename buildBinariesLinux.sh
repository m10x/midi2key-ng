#!/bin/bash
version=2.0.0

mkdir build/
rm build/*

# Linux amd64
goos=linux
goarch=amd64
GOOS=$goos GOARCH=$goarch go build -o midi2key-ng
zip build/midi2key-ng_"$version"_"$goos"_"$goarch".zip midi2key-ng

# Linux arm64
goos=linux
goarch=arm64
GOOS=$goos GOARCH=$goarch go build -o midi2key-ng
zip build/midi2key-ng_"$version"_"$goos"_"$goarch".zip midi2key-ng

# remove wcvs
rm midi2key-ng

# generate checksum file
find build/ -type f  \( -iname "*.zip" \) -exec sha256sum {} + > build/midi2key-ng_"$version"_checksums_sha256.txt