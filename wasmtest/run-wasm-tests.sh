#!/bin/bash
set -e
set -x

# install build tools
if ! which clang-9 ; then
	wget releases.llvm.org/9.0.0/clang+llvm-9.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz > /dev/null 2>&1
	tar xf clang+llvm-9.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz > /dev/null 2>&1
	export PATH="$(pwd)/clang+llvm-9.0.0-x86_64-linux-gnu-ubuntu-18.04/bin":$PATH
fi

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
