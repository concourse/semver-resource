#!/bin/bash

set -e

source $(dirname $0)/helpers.sh

it_can_check_with_no_current_version() {
  local repo=$(init_repo)

  check_uri $repo | jq -e "
    . == [{number: $(echo 0.0.0 | jq -R .)}]
  "
}

it_can_check_with_no_current_version_with_initial_set() {
  local repo=$(init_repo)

  check_uri_with_initial $repo 0.1.2 | jq -e "
    . == [{number: $(echo 0.1.2 | jq -R .)}]
  "
}

it_can_check_with_current_version() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  check_uri $repo | jq -e "
    . == [{number: $(echo 1.2.3 | jq -R .)}]
  "
}

it_fails_if_key_has_password() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  local key=$TMPDIR/key-with-passphrase
  ssh-keygen -f $key -N some-passphrase

  local failed_output=$TMPDIR/failed-output
  if check_uri_with_key $repo $key 2>$failed_output; then
    echo "checking should have failed"
    return 1
  fi

  grep "private keys with passphrases are not supported" $failed_output
}

it_can_check_with_credentials() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  check_uri_with_credentials $repo "user1" "pass1" | jq -e "
    . == [{number: $(echo 1.2.3 | jq -R .)}]
  "

  local expected_netrc="default login user1 password pass1"
  [ "$(cat $HOME/.netrc)" = "$expected_netrc" ]

  # make sure it clears out .netrc for this request without credentials
  check_uri_with_credentials $repo "" "" | jq -e "
    . == [{number: $(echo 1.2.3 | jq -R .)}]
  "
  [ ! -f "$HOME/.netrc" ]
}

it_clears_netrc_even_after_errors() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  if check_uri_with_credentials "non_existent_repo" "user1" "pass1" ; then
    exit 1
  fi

  local expected_netrc="default login user1 password pass1"
  [ "$(cat $HOME/.netrc)" = "$expected_netrc" ]

  # make sure it clears out .netrc for this request without credentials
  if check_uri_with_credentials "non_existent_repo" "" "" ; then
    exit 1
  fi
  [ ! -f "$HOME/.netrc" ]
}

it_can_check_from_a_version() {
  local repo=$(init_repo)

  set_version $repo 1.2.3

  check_uri_from $repo 1.2.3 | jq -e "
    . == [
      {number: $(echo 1.2.3 | jq -R .)}
    ]
  "

  check_uri_from $repo 1.2.4 | jq -e "
     . == [
      {number: $(echo 1.2.3 | jq -R .)}
    ]
  "

  set_version $repo 1.2.5

  check_uri_from $repo 1.2.4 | jq -e "
    . == [
      {number: $(echo 1.2.5 | jq -R .)}
    ]
  "

  set_version $repo 2.0.0

  check_uri_from $repo 1.2.4 | jq -e "
    . == [
      {number: $(echo 2.0.0 | jq -R .)}
    ]
  "
}

it_can_check_with_custom_file_location() {
  local repo=$(init_repo)
  local git_repo_path=$TMPDIR/semver-git-repo

  # Test with nested directory
  mkdir -p $repo/config
  echo "2.3.4" > $repo/config/version.txt
  git -C $repo add .
  git -C $repo \
    -c user.name='test' \
    -c user.email='test@example.com' \
    commit -q -m "add version"

  # Clean any existing repo
  rm -rf $git_repo_path

  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $repo | jq -R .),
      branch: \"master\",
      file: \"config/version.txt\"
    }
  }" | ${resource_dir}/check | jq -e "
    . == [{number: \"2.3.4\"}]
  "

  # Test with root file
  local repo2=$(init_repo)
  echo "1.0.0" > $repo2/VERSION
  git -C $repo2 add .
  git -C $repo2 \
    -c user.name='test' \
    -c user.email='test@example.com' \
    commit -q -m "add version"

  # MUST clean the cached repo before second test
  rm -rf $git_repo_path

  check_uri_with_file $repo2 VERSION | jq -e "
    . == [{number: \"1.0.0\"}]
  "
}

run it_can_check_with_no_current_version
run it_can_check_with_no_current_version_with_initial_set
run it_can_check_with_current_version
run it_fails_if_key_has_password
run it_can_check_with_credentials
run it_can_check_from_a_version
run it_clears_netrc_even_after_errors
run it_can_check_with_custom_file_location