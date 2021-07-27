#!/usr/bin/env bash

OS=$1
ARCH=$2
VERSION=$3

DEST=build/${OS}/${ARCH}
TAR=ciac.${VERSION}.${OS}-${ARCH}.tar.gz
BIN=ciac

cd `dirname $0`

CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -o ${DEST}/${BIN} ./cmd/ciac
cd ${DEST}
tar -czf ${TAR} ${BIN}
rm ${BIN}