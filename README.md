# Semver Resource

A resource for managing a version number. Persists the version number in an
S3 bucket.


## Source Configuration

* `initial_version`: *Optional.* The version number to use when
bootstrapping, i.e. when there is not a version number present in the bucket.

* `bucket`: *Required.* The name of the bucket.

* `key`: *Required.* The key to use for the object in the bucket tracking
the version.

* `access_key_id`: *Required.* The AWS access key to use when accessing the
bucket.

* `secret_access_key`: *Required.* The AWS secret key to use when accessing
the bucket.

* `region_name`: *Optional. Default `us-east-1`.* The region the bucket is in.

* `endpoint`: *Optional. Custom endpoint for using S3 compatible provider.

## Behavior

### `check`: Report the current version number.

Fetches the version number from the bucket and returns it as the resource's
version.


### `in`: Fetch the version number from the bucket, optionally bumping it.

Provides the version number to the build as a `number` file in the destination.

#### Parameters

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

Note that `bump` and `pre` don't update the version resource - they just
modify the version that gets provided to the build. An output must be
explicitly specified to actually update the version.


### `out`: Set the version to a number in a file.

Given a file, use its contents to update the version.

This is often used in combination with `in`. For example, the input bumps
the version number and provides it as `input-name/number`, and then the
build uses it, and then the output sets it as the new version.

#### Parameters

* `file`: *Required.* The file containing the version number.
