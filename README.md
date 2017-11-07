# Semver Resource

A resource for managing a version number. Persists the version number in one of several backing stores.


## Source Configuration

* `initial_version`: *Optional.* The version number to use when
bootstrapping, i.e. when there is not a version number present in the source.

* `driver`: *Optional. Default `s3`.* The driver to use for tracking the
  version. Determines where the version is stored.

There are three supported drivers, with their own sets of properties for
configuring them.


### `git` Driver

The `git` driver works by modifying a file in a repository with every bump. The
`git` driver has the advantage of being able to do atomic updates.

* `uri`: *Required.* The repository URL.

* `branch`: *Required.* The branch the file lives on.

* `file`: *Required.* The name of the file in the repository.

* `private_key`: *Optional.* The SSH private key to use when pulling from/pushing to to the repository.

* `username`: *Optional.* Username for HTTP(S) auth when pulling/pushing.
   This is needed when only HTTP/HTTPS protocol for git is available (which does not support private key auth)
   and auth is required.

* `password`: *Optional.* Password for HTTP(S) auth when pulling/pushing.

* `git_user`: *Optional.* The git identity to use when pushing to the
  repository support RFC 5322 address of the form "Gogh Fir \<gf@example.com\>" or "foo@example.com".

### `s3` Driver

The `s3` driver works by modifying a file in an S3 compatible bucket.

* `bucket`: *Required.* The name of the bucket.

* `key`: *Required.* The key to use for the object in the bucket tracking
the version.

* `access_key_id`: *Required.* The AWS access key to use when accessing the
bucket.

* `secret_access_key`: *Required.* The AWS secret key to use when accessing
the bucket.

* `region_name`: *Optional. Default `us-east-1`.* The region the bucket is in.

* `endpoint`: *Optional.* Custom endpoint for using S3 compatible provider.

* `disable_ssl`: *Optional.* Disable SSL for the endpoint, useful for S3 compatible providers without SSL.

* `skip_ssl_verification`: *Optional.* Skip SSL verification for S3 endpoint. Useful for S3 compatible providers using self-signed SSL certificates.

* `server_side_encryption`: *Optional.* The server-side encryption algorithm
used when storing the version object (e.g. `AES256`, `aws:kms`).

* `use_v2_signing`: *Optional.* Use v2 signing, default false.

### `swift` Driver

The `swift` driver works by modifying a file in a container.

* `openstack` *Required.* All openstack configuration must go under this key.

  * `container`: *Required.* The name of the container.

  * `item_name`: *Required.* The item name to use for the object in the container tracking
the version.

  * `region_name`: *Required.* The region the container is in.

  * `identity_endpoint`, `username`, `user_id`, `password`, `api_key`, `domain_id`, `domain_name`, `tenant_id`, `tenant_name`, `allow_reauth`, `token_id`: See below
The swift driver uses [gophercloud](http://gophercloud.io/docs/) to handle interacting
with OpenStack. All OpenStack Identity versions are supported through this library. The
Authentication properties will pass through to it. For detailed information about the
individual parameters, see https://github.com/rackspace/gophercloud/blob/master/auth_options.go

### `gcs` Driver

The `gcs` driver works by modifying a file in a Google Cloud Storage bucket.

* `bucket`: *Required.* The name of the bucket.

* `key`: *Required.* The key to use for the object in the bucket tracking the version.

* `json_key`: *Required.* The contents of your GCP Account JSON Key. Example:

  ```yaml
  json_key: |
    {
      "private_key_id": "...",
      "private_key": "...",
      "client_email": "...",
      "client_id": "...",
      "type": "service_account"
    }
  ```

### Example

With the following resource configuration:

``` yaml
resources:
- name: version
  type: semver
  source:
    driver: git
    uri: git@github.com:concourse/concourse.git
    branch: version
    file: version
    private_key: {{concourse-repo-private-key}}
```

Bumping with a `get` and then a `put`:

``` yaml
plan:
- get: version
  params: {bump: minor}
- task: a-thing-that-needs-a-version
- put: version
  params: {file: version/version}
```

Or, bumping with an atomic `put`:

``` yaml
plan:
- put: version
  params: {bump: minor}
- task: a-thing-that-needs-a-version
```

## Behavior

### `check`: Report the current version number.

Detects new versions by reading the file from the specified source. If the file is empty, it returns the `initial_version`. If the file is not empty, it returns the version specified in the file if it is equal to or greater than current version, otherwise it returns no versions.

### `in`: Provide the version as a file, optionally bumping it.

Provides the version number to the build as a `version` file in the destination.

Can be configured to bump the version locally, which can be useful for getting
the `final` version ahead of time when building artifacts.

#### Parameters

* `bump` and `pre`: *Optional.* See [Version Bumping
  Semantics](#version-bumping-semantics).

Note that `bump` and `pre` don't update the version resource - they just
modify the version that gets provided to the build. An output must be
explicitly specified to actually update the version.


### `out`: Set the version or bump the current one.

Given a file, use its contents to update the version. Or, given a bump
strategy, bump whatever the current version is. If there is no current version,
the bump will be based on `initial_version`.

The `file` parameter should be used if you have a particular version that you
want to force the current version to be. This can be used in combination with
`in`, but it's probably better to use the `bump` and `pre` params as they'll
perform an atomic in-place bump if possible with the driver.

#### Parameters

One of the following must be specified:

* `file`: *Optional.* Path to a file containing the version number to set.

* `bump` and `pre`: *Optional.* See [Version Bumping
  Semantics](#version-bumping-semantics).

When `bump` and/or `pre` are used, the version bump will be applied atomically,
if the driver supports it. That is, if we pull down version `N`, and bump to
`N+1`, the driver can then compare-and-swap. If the compare-and-swap fails
because there's some new version `M`, the driver will re-apply the bump to get
`M+1`, and try again (in a loop).


## Version Bumping Semantics

Both `in` and `out` support bumping the version semantically via two params:
`bump` and `pre`:

* `bump`: *Optional.* Bump the version number semantically. The value must
be one of:

  * `major`: Bump the major version number, e.g. `1.0.0` -> `2.0.0`.
  * `minor`: Bump the minor version number, e.g. `0.1.0` -> `0.2.0`.
  * `patch`: Bump the patch version number, e.g. `0.0.1` -> `0.0.2`.
  * `final`: Promote the version to a final version, e.g. `1.0.0-rc.1` -> `1.0.0`.


* `pre`: *Optional.* When bumping, bump to a prerelease (e.g. `rc` or
`alpha`), or bump an existing prerelease.

  If present, and the version is already a prerelease matching this value,
  its number is bumped. If the version is already a prerelease of another
  type, (e.g. `alpha` vs. `beta`), the type is switched and the prerelease
  version is reset to `1`. If the version is *not* already a pre-release, then
  `pre` is added, starting at `1`.

### Running the tests

The tests have been embedded with the `Dockerfile`; ensuring that the testing
environment is consistent across any `docker` enabled platform. When the docker
image builds, the test are run inside the docker container, on failure they
will stop the build.

Run the tests with the following command:

```sh
docker build -t semver-resource .
```

#### Integration tests

The integration requires two AWS S3 buckets, one without versioning and another
with. The `docker build` step requires setting `--build-args` so the
integration will run.

You will need:
* AWS key and secret
* An S3 bucket
* The region you are in (i.e. `us-east-1`, `us-west-2`)

Run the tests with the following command, replacing each `build-arg` value with your own values:

```sh
docker build . -t semver-resource --build-arg SEMVER_TESTING_ACCESS_KEY_ID="some-key" --build-arg SEMVER_TESTING_SECRET_ACCESS_KEY="some-secret" --build-arg SEMVER_TESTING_BUCKET="some-bucket" --build-arg SEMVER_TESTING_REGION="some-region"
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
