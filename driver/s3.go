package driver

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/mitchellh/goamz/s3"
)

type S3Driver struct {
	Bucket *s3.Bucket
	Key    string
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
		return []semver.Version{}, nil
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
