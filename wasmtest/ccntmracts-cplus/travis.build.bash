#!/usr/bin/env bash
set -ev

oldir=$(pwd)
currentdir=$(dirname $0)
cd $currentdir

git clone --recursive https://github.com/cntmio/cntmology-wasm-cdt-cpp
cd cntmology-wasm-cdt-cpp; git checkout v1.0 -b testframe; bash compiler_install.bash;cd ../
compilerdir="./cntmology-wasm-cdt-cpp/install/bin"

for f in $(ls *.cpp)
do
	$compilerdir/cntm_cpp $f -lbase58 -lbuiltins -o  ${f%.cpp}.wasm
done

rm -rf cntmology-wasm-cdt-cpp
mv *.wasm ../testwasmdata/
rm *.wasm.str
cp  *.avm ../testwasmdata/

cd $oldir
