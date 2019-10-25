#!/bin/bash

mkdir -p dist

docker run \
    -e GOOS=darwin -e GOARCH=amd64 \
    -v $(pwd)/:/go/src/github.com/anonhoarder/demeter \
    -w /go/src/github.com/anonhoarder/demeter \
    golang go build -ldflags="-s -w" -o dist/darwin_amd64_demeter .

docker run \
    -e GOOS=linux -e GOARCH=amd64 \
    -v $(pwd)/:/go/src/github.com/anonhoarder/demeter \
    -w /go/src/github.com/anonhoarder/demeter \
    golang go build -ldflags="-s -w" -o dist/linux_amd64_demeter .

docker run \
    -e GOOS=linux -e GOARCH=arm -e GOARM=5  \
    -v $(pwd)/:/go/src/github.com/anonhoarder/demeter \
    -w /go/src/github.com/anonhoarder/demeter \
    golang go build -ldflags="-s -w" -o dist/linux_arm_demeter .
