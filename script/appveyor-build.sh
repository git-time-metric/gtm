#!/bin/bash

set -ex

export PATH=/c/msys64/mingw64/bin:/c/msys64/usr/bin:/c/Go/bin:/c/gopath/go/bin:$PATH

GIT2GO_PATH="${GOPATH}/src/github.com/libgit2/git2go"
LIBGIT2_BUILD="${GIT2GO_PATH}/vendor/libgit2/build"
mkdir -p "${LIBGIT2_BUILD}"
cd "${LIBGIT2_BUILD}"
FLAGS="-lws2_32"
export CGO_LDFLAGS="${LIBGIT2_BUILD}/libgit2.a -L${LIBGIT2_BUILD} ${FLAGS}"

cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX=../install \
      -DWINHTTP=OFF \
      -DUSE_BUNDLED_ZLIB=ON \
      -DUSE_HTTPS=OFF \
      -DUSE_SSH=OFF \
      -DCURL=OFF \
      -G "MSYS Makefiles" \
      .. &&
cmake --build .

cd "${GOPATH}/src/github.com/git-time-metric/gtm"
go get -d ./...
go test --tags static $(go list ./... | grep -v vendor)
if [[ "${APPVEYOR_REPO_TAG}" = true ]]; then
    version=${APPVEYOR_REPO_TAG_NAME}
    go build -v --tags static -ldflags "-X main.Version=${version}"

    # make sure version is set correctly
    v="$(${GOPATH}/src/github.com/kilpkonn/gtm-enhanced/gtm.exe verify ${version})"
    if [ ! "$v" == "true" ]; then
        exit 1
    fi

    tar -zcf "gtm.${APPVEYOR_REPO_TAG_NAME}.windows.tar.gz" gtm.exe
else
    version='0.0.0-dev'
    go build -v --tags static -ldflags "-X main.Version=${version}"

    # make sure version is set correctly
    v="$(${GOPATH}/src/github.com/kilpkonn/gtm-enhanced/gtm.exe verify ${version})"
    if [ ! "$v" == "true" ]; then
        exit 1
    fi

    tar -zcf "gtm.developer-build-${version}.windows.tar.gz" gtm.exe
fi
