#!/bin/sh

set -e -u

set -o pipefail

resource_dir=/opt/resource

run() {
  export TMPDIR=$(mktemp -d ${TMPDIR_ROOT}/git-tests.XXXXXX)

  echo -e 'running \e[33m'"$@"$'\e[0m...'
  eval "$@" 2>&1 | sed -e 's/^/  /g'
  echo ""
}

init_repo() {
  (
    set -e

    cd $(mktemp -d $TMPDIR/repo.XXXXXX)

    git init -q

    # start with an initial commit
    git \
      -c user.name='test' \
      -c user.email='test@example.com' \
      commit -q --allow-empty -m "init"

    # print resulting repo
    pwd
  )
}

set_version_in_file_on_branch() {
  local repo=$1
  local file=$2
  local branch=$3
  local version=${4}

  # ensure branch exists
  if ! git -C $repo rev-parse --verify $branch >/dev/null; then
    git -C $repo branch $branch master
  fi

  # switch to branch
  git -C $repo checkout -q $branch

  # modify file and commit
  echo $version > $repo/$file
  git -C $repo add $file
  git -C $repo \
    -c user.name='test' \
    -c user.email='test@example.com' \
    commit -q -m "set version to $version"

  # output resulting sha
  git -C $repo rev-parse HEAD
}

set_version() {
  set_version_in_file_on_branch $1 some-file master $2
}

check_uri() {
  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $1 | jq -R .),
      branch: \"master\",
      file: \"some-file\"
    }
  }" | ${resource_dir}/check | tee /dev/stderr
}

check_uri_with_initial() {
  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $1 | jq -R .),
      branch: \"master\",
      file: \"some-file\",
      initial_version: $(echo $2 | jq -R .)
    }
  }" | ${resource_dir}/check | tee /dev/stderr
}


check_uri_with_key() {
  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $1 | jq -R .),
      branch: \"master\",
      file: \"some-file\",
      private_key: $(cat $2 | jq -s -R .)
    }
  }" | ${resource_dir}/check | tee /dev/stderr
}


check_uri_with_credentials() {
  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $1 | jq -R .),
      branch: \"master\",
      file: \"some-file\",
      username: $(echo $2 | jq -R .),
      password: $(echo $3 | jq -R .)
    }
  }" | ${resource_dir}/check | tee /dev/stderr
}


check_uri_from() {
  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $1 | jq -R .),
      branch: \"master\",
      file: \"some-file\"
    },
    version: {
      number: $(echo $2 | jq -R .)
    }
  }" | ${resource_dir}/check | tee /dev/stderr
}

put_uri() {
  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $1 | jq -R .),
      branch: \"master\",
      file: \"some-file\"
    },
    params: {
      file: $(echo $3 | jq -R .)
    }
  }" | ${resource_dir}/out "$2" | tee /dev/stderr
}

put_uri_with_bump() {
  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $1 | jq -R .),
      branch: \"master\",
      file: \"some-file\"
    },
    params: {
      bump: $(echo $3 | jq -R .),
      pre: $(echo $4 | jq -R .)
    }
  }" | ${resource_dir}/out "$2" | tee /dev/stderr
}

put_uri_with_bump_and_initial() {
  jq -n "{
    source: {
      driver: \"git\",
      uri: $(echo $1 | jq -R .),
      branch: \"master\",
      file: \"some-file\",
      initial_version: $(echo $3 | jq -R .)
    },
    params: {
      bump: $(echo $4 | jq -R .),
      pre: $(echo $5 | jq -R .)
    }
  }" | ${resource_dir}/out "$2" | tee /dev/stderr
}
