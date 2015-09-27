# Semver Resource

A resource for managing a version number. Persists the version number in an
S3 bucket, though more storage drivers will be coming soon.


## Source Configuration

* `driver`: *Optional. Default `s3`.* The driver to use for tracking the
  version. Currently the only supported driver is `s3`, though `git` support is
  coming soon.

* `initial_version`: *Optional.* The version number to use when
bootstrapping, i.e. when there is not a version number present in the source.

* `bucket`: *Required.* The name of the bucket.

* `key`: *Required.* The key to use for the object in the bucket tracking
the version.

* `access_key_id`: *Required.* The AWS access key to use when accessing the
bucket.

* `secret_access_key`: *Required.* The AWS secret key to use when accessing
the bucket.

* `region_name`: *Optional. Default `us-east-1`.* The region the bucket is in.

* `endpoint`: *Optional.* Custom endpoint for using S3 compatible provider.


### Example

``` yaml
- name: version
  type: semver
  source:
    bucket: version-numbers
    key: product-version
    access_key_id: AKIANOTAKEY
    secret_access_key: li984n9sd0nfgns833rwwsd\s293
```

Bumping with a `get` and then a `put`:

``` yaml
- get: version
  params: {bump: minor}
- task: a-thing-that-needs-a-version
- put: version
  params: {file: version/number}
```

Or, bumping with an atomic `put`:

``` yaml
- put: version
  params: {bump: minor}
- task: a-thing-that-needs-a-version
```

## Behavior

### `check`: Report the current version number.

Detects new versions, currently by reading the file in S3 and either emitting
the initial version if it's not present, or emitting the current version if
it's newer than the current one.


### `in`: Provide the version as a file, optionally bumping it.

Provides the version number to the build as a `number` file in the destination.

#### Parameters

* `bump` and `pre`: *Optional.* See [Version Bumping
  Semantics](#version-bumping-semantics).

Note that `bump` and `pre` don't update the version resource - they just
modify the version that gets provided to the build. An output must be
explicitly specified to actually update the version.


### `out`: Set the version or bump the current one.

Given a file, use its contents to update the version. Or, given a bump strategy,
bump whatever the current version is.

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
version is reset to `1`. If the version is *not* already a pre-release,
`bump` is applied, and then `pre` is added, starting at `1`.

So, with `bump` as `minor` and `pre` as `rc`, the following changes will be
applied:

* `0.1.0` -> `0.2.0-rc.1`
* `0.1.0-rc.1` -> `0.1.0-rc.2`
* `0.1.0-beta.4` -> `0.1.0-rc.1`
