#!/bin/bash
tool=midi2key-ng
version=1.0.0
dir=build

mkdir build

# Windows amd64
goos=windows
goarch=amd64
set GOOS=$goos
set GOARCH=$goarch
go build -buildmode=exe -o "$dir"/"$tool".exe
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip "$dir"/"$tool".exe

# Linux amd64
goos=linux
goarch=amd64
set GOOS=$goos
set GOARCH=$goarch
go build -o "$dir"/"$tool"
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip "$dir"/"$tool"

# Linux arm64
goos=linux
goarch=arm64
set GOOS=$goos
set GOARCH=$goarch
go build -o "$dir"/"$tool"
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip "$dir"/"$tool"

# Darwin/MacOS amd64
goos=darwin
goarch=amd64
set GOOS=$goos
set GOARCH=$goarch
go build -o "$dir"/"$tool"
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip "$dir"/"$tool"

# Darwin/MacOS arm64
goos=darwin
goarch=arm64
set GOOS=$goos
set GOARCH=$goarch
go build -o "$dir"/"$tool"
zip build/"$tool"_"$version"_"$goos"_"$goarch".zip "$dir"/"$tool"

# reset GOOS and GOARCH
set GOOS=
set GOARCH=

# remove wcvs
rm "$dir"/"$tool"
rm "$dir"/"$tool".exe

# generate checksum file
find "$dir"/ -type f  \( -iname "*.zip" \) -exec sha256sum {} + > "$dir"/"$tool"_"$version"_checksums_sha256.txt