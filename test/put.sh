#!/bin/bash

set -e

source $(dirname $0)/helpers.sh

it_can_put_and_set_first_version() {
  local repo=$(init_repo)

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)
  echo 1.2.3 > $src/some-new-file

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri $repo $src some-new-file | jq -e "
    .version == {number: \"1.2.3\"}
  "

  # switch back to master
  git -C $repo checkout master

  test -e $repo/some-file
  test "$(cat $repo/some-file)" = 1.2.3
}

it_can_put_and_set_same_version() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)
  echo 1.2.3 > $src/some-new-file

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri $repo $src some-new-file | jq -e "
    .version == {number: \"1.2.3\"}
  "

  # switch back to master
  git -C $repo checkout master

  test -e $repo/some-file
  test "$(cat $repo/some-file)" = 1.2.3
}

it_can_put_and_set_over_existing_version() {
  local repo=$(init_repo)

  set_version $repo 0.0.1

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)
  echo 1.2.3 > $src/some-new-file

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri $repo $src some-new-file | jq -e "
    .version == {number: \"1.2.3\"}
  "

  # switch back to master
  git -C $repo checkout master

  test -e $repo/some-file
  test "$(cat $repo/some-file)" = 1.2.3
}

it_can_put_and_bump_first_version() {
  local repo=$(init_repo)

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri_with_bump $repo $src minor alpha | jq -e "
    .version == {number: \"0.1.0-alpha.1\"}
  "

  # switch back to master
  git -C $repo checkout master

  test -e $repo/some-file
  test "$(cat $repo/some-file)" = 0.1.0-alpha.1
}

it_can_put_and_bump_first_version_with_initial() {
  local repo=$(init_repo)

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri_with_bump_and_initial $repo $src 1.2.3 minor alpha | jq -e "
    .version == {number: \"1.3.0-alpha.1\"}
  "

  # switch back to master
  git -C $repo checkout master

  test -e $repo/some-file
  test "$(cat $repo/some-file)" = 1.3.0-alpha.1
}

it_can_put_and_bump_over_existing_version() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri_with_bump $repo $src minor alpha | jq -e "
    .version == {number: \"1.3.0-alpha.1\"}
  "

  # switch back to master
  git -C $repo checkout master

  test -e $repo/some-file
  test "$(cat $repo/some-file)" = 1.3.0-alpha.1
}

it_can_put_and_bump_with_message_over_existing_version() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)

  local message="This is a commit message"

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri_with_bump_and_message $repo $src minor alpha "$message" | jq -e "
    .version == {number: \"1.3.0-alpha.1\"}
  "

  # switch back to master
  git -C $repo checkout master

  test -e $repo/some-file
  test "$(cat $repo/some-file)" = 1.3.0-alpha.1
  test "$(git -C $repo log -n1 --pretty=%B)" = "$message"
}

it_can_put_and_bump_with_message_and_replace_over_existing_version() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)

  local message="This is a commit message on %file% with %version%"

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri_with_bump_and_message $repo $src minor alpha "$message" | jq -e "
    .version == {number: \"1.3.0-alpha.1\"}
  "

  # switch back to master
  git -C $repo checkout master

  local expected_message="This is a commit message on some-file with 1.3.0-alpha.1"

  test -e $repo/some-file
  test "$(cat $repo/some-file)" = 1.3.0-alpha.1
  test "$(git -C $repo log -n1 --pretty=%B)" = "$expected_message"
}

run it_can_put_and_set_first_version
run it_can_put_and_set_same_version
run it_can_put_and_set_over_existing_version
run it_can_put_and_bump_first_version
run it_can_put_and_bump_first_version_with_initial
run it_can_put_and_bump_over_existing_version
run it_can_put_and_bump_with_message_over_existing_version
run it_can_put_and_bump_with_message_and_replace_over_existing_version
