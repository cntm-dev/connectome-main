#!/usr/bin/env bash
set -ex

VERSION=$(git describe --always --tags --lcntm)

if [ $TRAVIS_OS_NAME == 'linux' ]; then
	echo "linux sys"
	env GO111MODULE=on make all
	cd ./wasmtest && bash ./run-wasm-tests.sh && cd ../
	bash ./.travis.check-license.sh
	bash ./.travis.check-templog.sh
	bash ./.travis.gofmt.sh
	bash ./.travis.gotest.sh
elif [ $TRAVIS_OS_NAME == 'osx' ]; then
	echo "osx sys"
	env GO111MODULE=on make all
else
	echo "win sys"
	env GO111MODULE=on CGO_ENABLED=1 go build  -ldflags "-X github.com/cntmio/cntmology/common/config.Version=${VERSION}" -o cntmology-windows-amd64 main.go
	env GO111MODULE=on go build  -ldflags "-X github.com/cntmio/cntmology/common/config.Version=${VERSION}" -o sigsvr-windows-amd64 sigsvr.go
fi
