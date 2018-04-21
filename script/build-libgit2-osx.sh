#!/bin/bash

set -ex

pushd ${PWD}

cd ${GOPATH}/src/github.com/libgit2/git2go/vendor/libgit2 &&
mkdir -p install/lib &&
mkdir -p build &&
cd build &&
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DUSE_SSH=OFF \
      -DCURL=OFF \
      -DUSE_HTTPS=OFF \
      -DUSE_BUNDLED_ZLIB=ON \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX=../install \
      .. &&

cmake --build .

popd
