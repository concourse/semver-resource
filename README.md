# Semver Resource

A resource for managing a version number. Persists the version number in an
S3 bucket or a Git repository.


## Source Configuration

* `initial_version`: *Optional.* The version number to use when
bootstrapping, i.e. when there is not a version number present in the source.

* `driver`: *Optional. Default `s3`.* The driver to use for tracking the
  version. Determines where the version is stored.

There are two supported drivers, with their own sets of properties for
configuring them.


### `git` Driver

The `git` driver works by modifying a file in a repository with every bump. The
`git` driver has the advantage of being able to do atomic updates.

* `uri`: *Required.* The repository URL.

* `branch`: *Required.* The branch the file lives on.

* `file`: *Required.* The name of the file in the repository.

* `private_key`: *Optional.* The SSH private key to use when pulling
  from/pushing to to the repository.


### `s3` Driver

The `s3` driver works by modifying a file a file in a bucket.

* `bucket`: *Required.* The name of the bucket.

* `key`: *Required.* The key to use for the object in the bucket tracking
the version.

* `access_key_id`: *Optional.* The AWS access key to use when accessing the
bucket. If empty, the resource will try to retrieve credentials from environment
or [AWS instance profile](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html)

* `secret_access_key`: *Optional.* The AWS secret key to use when accessing
the bucket. If empty, the resource will try to retrieve credentials from environment
or [AWS instance profile](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html)

* `region_name`: *Optional. Default `us-east-1`.* The region the bucket is in.

* `endpoint`: *Optional.* Custom endpoint for using S3 compatible provider.


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
  params: {file: version/number}
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

Detects new versions, currently by reading the file from the source and either
emitting the initial version if it's not present, or emitting the current
version if it's newer than the current one.


### `in`: Provide the version as a file, optionally bumping it.

Provides the version number to the build as a `number` file in the destination.

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
