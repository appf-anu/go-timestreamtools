#!/bin/bash
result=${PWD##*/}
export GOARCH=amd64
fn="${1:-$result}"
filename=$(basename "$fn")
extension="${filename##*.}"
filename="${filename%.*}"
env GOOS=windows go test "$1"
env GOOS=lunix go test "$1"
env GOOS=darwin go test "$1"
env GOOS=windows go build -o "$filename"_win-"$GOARCH".exe "$1"
env GOOS=linux go build -o "$filename"_linux-"$GOARCH" "$1"
env GOOS=darwin go build -o "$filename"_darwin-"$GOARCH" "$1"
