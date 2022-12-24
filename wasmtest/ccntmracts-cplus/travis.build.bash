#!/usr/bin/env bash
set -ev

oldir=$(pwd)
currentdir=$(dirname $0)
cd $currentdir

git clone https://github.com/cntmio/cntmology-wasm-cdt-cpp
compilerdir="./cntmology-wasm-cdt-cpp/install/bin"

for f in $(ls *.cpp)
do
	$compilerdir/cntm_cpp $f -lbase58 -lcrypto -lbuiltins -o  ${f%.cpp}.wasm
done

rm -rf cntmology-wasm-cdt-cpp
mv *.wasm ../testwasmdata/
rm *.wasm.str
cp  *.avm ../testwasmdata/

cd $oldir
