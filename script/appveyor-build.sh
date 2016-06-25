#!/bin/sh
set -ex

export PATH=/c/msys64/mingw64/bin:/c/msys64/usr/bin:/c/Go/bin:/c/gopath/go/bin:$PATH
export GOROOT=/c/Go/
export GOPATH=/c/gopath

pacman -S --noconfirm mingw-w64-x86_64-libssh2 

go get -d github.com/libgit2/git2go
cd /c/gopath/src/github.com/libgit2/git2go
git checkout next
git submodule update --init

cp /c/gopath/src/github.com/git-time-metric/gtm/script/build-libgit2-static.sh \
   /c/gopath/src/github.com/libgit2/git2go/script/build-libgit2-static.sh

make install

cd /c/gopath/src/github.com/git-time-metric/gtm
go get -t -v ./...
go test -v ./...
if [[ "${APPVEYOR_REPO_TAG}" = true ]]; then
  go build -v -ldflags "-X main.version=${APPVEYOR_REPO_TAG_NAME}"
  7z a gtm.${APPVEYOR_REPO_TAG}.w64-x86_64.zip ${APPVEYOR_BUILD_FOLDER}/gtm.exe
fi
