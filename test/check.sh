#!/bin/sh

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
  [ "$(cat $HOME/.netrc)" = "$expected_netrc" ] && rm -f $HOME/.netrc
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
    . == []
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

run it_can_check_with_no_current_version
run it_can_check_with_no_current_version_with_initial_set
run it_can_check_with_current_version
run it_fails_if_key_has_password
run it_can_check_with_credentials
run it_can_check_from_a_version
