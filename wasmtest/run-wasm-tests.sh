#!/bin/bash
set -e
set -x

# install build tools
if ! which rustup ; then
	curl https://sh.rustup.rs -sSf | sh -s -- -y --default-toolchain nightly 
	source $HOME/.cargo/env
fi
rustup target add wasm32-unknown-unknown
which cntmio-wasm-build || cargo install --git=https://github.com/cntmio/cntmio-wasm-build

# build rust wasm ccntmracts
mkdir -p testwasmdata
cd ccntmracts-rust && bash travis.build.sh && cd ../

cd ccntmracts-cplus && bash travis.build.bash && cd ../

# verify and optimize wasm ccntmract
for wasm in testwasmdata/*.wasm ; do
	cntmio-wasm-build $wasm $wasm
done

# start test framework
go run wasm-test.go
