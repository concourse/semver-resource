#!/bin/sh

set -e

source $(dirname $0)/helpers.sh

it_can_put_and_set_first_version() {
  local driver=$1
  local repo=$(init_repo)

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)
  echo 1.2.3 > $src/some-new-file

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri $driver $repo $src some-new-file | jq -e "
    .version == {number: \"1.2.3\"}
  "

  # switch back to master
  git -C $repo checkout master

  check_version $driver $repo 1.2.3
}

it_can_put_and_set_same_version() {
  local driver=$1
  local repo=$(init_repo)

  set_version $repo 1.2.3

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)
  echo 1.2.3 > $src/some-new-file

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri $driver $repo $src some-new-file | jq -e "
    .version == {number: \"1.2.3\"}
  "

  # switch back to master
  git -C $repo checkout master

  check_version $driver $repo 1.2.3
}

it_can_put_and_set_over_existing_version() {
  local driver=$1
  local repo=$(init_repo)

  set_version $repo 0.0.1
  sleep 1 # to ensure the next tag has a different timestamp

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)
  echo 1.2.3 > $src/some-new-file

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri $driver $repo $src some-new-file | jq -e "
    .version == {number: \"1.2.3\"}
  "

  # switch back to master
  git -C $repo checkout master

  check_version $driver $repo 1.2.3
}

it_can_put_and_bump_first_version() {
  local driver=$1
  local repo=$(init_repo)

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri_with_bump $driver $repo $src minor alpha | jq -e "
    .version == {number: \"0.1.0-alpha.1\"}
  "

  # switch back to master
  git -C $repo checkout master

  check_version $driver $repo 0.1.0-alpha.1
}

it_can_put_and_bump_first_version_with_initial() {
  local driver=$1
  local repo=$(init_repo)

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri_with_bump_and_initial $driver $repo $src 1.2.3 minor alpha | jq -e "
    .version == {number: \"1.3.0-alpha.1\"}
  "

  # switch back to master
  git -C $repo checkout master

  check_version $driver $repo 1.3.0-alpha.1
}

it_can_put_and_bump_over_existing_version() {
  local driver=$1
  local repo=$(init_repo)

  set_version $repo 1.2.3
  sleep 1 # to ensure the next tag has a different timestamp

  local src=$(mktemp -d $TMPDIR/put-src.XXXXXX)

  # cannot push to repo while it's checked out to a branch
  git -C $repo checkout refs/heads/master

  put_uri_with_bump $driver $repo $src minor alpha | jq -e "
    .version == {number: \"1.3.0-alpha.1\"}
  "

  # switch back to master
  git -C $repo checkout master

  check_version $driver $repo 1.3.0-alpha.1
}

run it_can_put_and_set_first_version $1
run it_can_put_and_set_same_version $1
run it_can_put_and_set_over_existing_version $1
run it_can_put_and_bump_first_version $1
run it_can_put_and_bump_first_version_with_initial $1
run it_can_put_and_bump_over_existing_version $1
