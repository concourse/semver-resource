package driver

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
	"github.com/mitchellh/goamz/s3"
)

type S3Driver struct {
	InitialVersion semver.Version

	Bucket *s3.Bucket
	Key    string
}

func (driver *S3Driver) Bump(bump version.Bump) (semver.Version, error) {
	var currentVersion semver.Version

	bucketNumberPayload, err := driver.Bucket.Get(driver.Key)
	if err == nil {
		currentVersion, err = semver.Parse(string(bucketNumberPayload))
		if err != nil {
			return semver.Version{}, err
		}
	} else if s3err, ok := err.(*s3.Error); ok && s3err.StatusCode == 404 {
		currentVersion = driver.InitialVersion
	} else {
		return semver.Version{}, err
	}

	newVersion := bump.Apply(currentVersion)

	err = driver.Set(newVersion)
	if err != nil {
		return semver.Version{}, err
	}

	return newVersion, nil
}

func (driver *S3Driver) Set(newVersion semver.Version) error {
	return driver.Bucket.Put(driver.Key, []byte(newVersion.String()), "text/plain", s3.Private)
}

func (driver *S3Driver) Check(cursor *semver.Version) ([]semver.Version, error) {
	var bucketNumber string

	bucketNumberPayload, err := driver.Bucket.Get(driver.Key)
	if err == nil {
		bucketNumber = string(bucketNumberPayload)
	} else if s3err, ok := err.(*s3.Error); ok && s3err.StatusCode == 404 {
		if cursor == nil {
			return []semver.Version{driver.InitialVersion}, nil
		} else {
			return []semver.Version{}, nil
		}
	} else {
		return nil, err
	}

	bucketVersion, err := semver.Parse(bucketNumber)
	if err != nil {
		return nil, fmt.Errorf("parsing number in bucket: %s", err)
	}

	if cursor == nil || bucketVersion.GT(*cursor) {
		return []semver.Version{bucketVersion}, nil
	}

	return []semver.Version{}, nil
}
