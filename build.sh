#!/usr/bin/env bash

export VERSION=$(git -C "$HELM_PLUGIN_DIR" describe --tags --exact-match 2>/dev/null || :)
export CURDIR=`pwd`
export PKG=github.com/cic-sap/helmize

dist(){

  set -ex
  export LDFLAGS=" -X ${PKG}/cmd.Version=${VERSION}"


  export CGO_ENABLED=0
  mkdir -p build/helmize/bin release/
	rm -rf build/helmize/* release/*
	mkdir -p build/helmize/bin release/
	cp README.md LICENSE plugin.yaml build/helmize
	GOOS=linux GOARCH=amd64 go build -o build/helmize/bin/helmize -trimpath -ldflags="${LDFLAGS}"
	tar -C build/ -zcvf ${CURDIR}/release/helmize-linux.amd64.tgz helmize/
	GOOS=linux GOARCH=arm64 go build -o build/helmize/bin/helmize -trimpath -ldflags="${LDFLAGS}"
	tar -C build/ -zcvf ${CURDIR}/release/helmize-linux.arm64.tgz helmize/

	GOOS=freebsd GOARCH=amd64 go build -o build/helmize/bin/helmize -trimpath -ldflags="${LDFLAGS}"
	tar -C build/ -zcvf ${CURDIR}/release/helmize-freebsd.tgz helmize/
	GOOS=darwin GOARCH=amd64 go build -o build/helmize/bin/helmize -trimpath -ldflags="${LDFLAGS}"
	tar -C build/ -zcvf ${CURDIR}/release/helmize-macos.amd64.tgz helmize/

  GOOS=darwin GOARCH=arm64 go build -o build/helmize/bin/helmize -trimpath -ldflags="${LDFLAGS}"
	tar -C build/ -zcvf ${CURDIR}/release/helmize-macos.arm64.tgz helmize/

	rm build/helmize/bin/helmize
	GOOS=windows GOARCH=amd64 go build -o build/helmize/bin/helmize.exe -trimpath -ldflags="${LDFLAGS}"
	tar -C build/ -zcvf ${CURDIR}/release/helmize-windows.tgz helmize/
}
dist
