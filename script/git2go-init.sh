#!/bin/sh
set -ex

COMMIT_HASH=v26
PROJPATH="$GOPATH/src/github.com/libgit2/git2go"
[ ! -f "$GOPATH/src/github.com/libgit2/git2go/.git/config" ] && git clone https://github.com/libgit2/git2go.git $PROJPATH
cd $PROJPATH
git checkout -qf $COMMIT_HASH
git submodule update --init
sed -i -- 's/ZLIB_FOUND/FALSE/g' $PROJPATH/vendor/libgit2/CMakeLists.txt
sed -i -- 's/OPENSSL_FOUND/FALSE/g' $PROJPATH/vendor/libgit2/CMakeLists.txt
sed -i -- 's/USE_SSH.*"Link with libssh to enable SSH support".*ON/USE_SSH  "Link with libssh to enable SSH support"  OFF/g' $PROJPATH/vendor/libgit2/CMakeLists.txt
