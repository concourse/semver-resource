package driver

import (
	"errors"
	"fmt"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type Driver interface {
	Set(semver.Version) error
	Check(*semver.Version) ([]semver.Version, error)
}

func FromSource(source models.Source) (Driver, error) {
	switch source.Driver {
	case models.DriverUnspecified, models.DriverS3:
		auth := aws.Auth{
			AccessKey: source.AccessKeyID,
			SecretKey: source.SecretAccessKey,
		}

		regionName := source.RegionName
		if len(regionName) == 0 {
			regionName = aws.USEast.Name
		}

		region, ok := aws.Regions[regionName]
		if !ok {
			return nil, errors.New(fmt.Sprintf("no such region '%s'", regionName))
		}

		if len(source.Endpoint) != 0 {
			region = aws.Region{
				S3Endpoint: fmt.Sprintf("https://%s", source.Endpoint),
			}
		}

		client := s3.New(auth, region)
		bucket := client.Bucket(source.Bucket)

		return &S3Driver{
			Bucket: bucket,
			Key:    source.Key,
		}, nil

	default:
		return nil, fmt.Errorf("unknown driver: %s", source.Driver)
	}
}
