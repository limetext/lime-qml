#!/usr/bin/env bash

source "$(dirname -- "$0")/../../general/ci/setup.sh"

ret=0

fold_start "build" "build"
build "main"
let ret=$ret+$build_result
fold_end "build"

exit $ret
