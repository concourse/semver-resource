#!/bin/sh

set -e

export TMPDIR_ROOT=$(mktemp -d /tmp/git-tests.XXXXXX)

$(dirname $0)/check.sh "git"
$(dirname $0)/put.sh "git"
rm -rf $TMPDIR_ROOT/*
$(dirname $0)/check.sh "git-tag"
$(dirname $0)/put.sh "git-tag"

echo -e '\e[32mall tests passed!\e[0m'

rm -rf $TMPDIR_ROOT
