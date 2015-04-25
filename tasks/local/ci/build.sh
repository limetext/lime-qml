#!/usr/bin/env bash

source "$(dirname -- "$0")/../../general/ci/setup.sh"

fold_start "get" "get dependencies"
go get github.com/limetext/lime-qml/main/...
fold_end "get"

ret=0

fold_start "build" "build"
build "main"
let ret=$ret+$build_result
fold_end "build"

exit $ret
