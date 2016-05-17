#!/bin/bash

mkdir -p /go
export GOPATH=/go

mkdir -p /go/src/github.com/kopeio/
ln -s /src /go/src/github.com/kopeio/aws-es-proxy

cd /go/src/github.com/kopeio/aws-es-proxy
/usr/bin/glide install

go install .

mkdir -p /src/.build/artifacts/
cp /go/bin/aws-es-proxy /src/.build/artifacts/
