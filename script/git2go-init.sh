#!/bin/bash

set -ex

pushd ${PWD}

COMMIT_HASH=v27
PROJPATH="$GOPATH/src/github.com/libgit2/git2go"
[ ! -f "$GOPATH/src/github.com/libgit2/git2go/.git/config" ] && git clone https://github.com/libgit2/git2go.git $PROJPATH
cd $PROJPATH
git checkout -qf $COMMIT_HASH
git submodule update --init

popd
