package driver

import (
	"errors"
	"fmt"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
	"github.com/concourse/semver-resource/version"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type Driver interface {
	Bump(version.Bump) (semver.Version, error)
	Set(semver.Version) error
	Check(*semver.Version) ([]semver.Version, error)
}

func FromSource(source models.Source) (Driver, error) {
	var initialVersion semver.Version
	if source.InitialVersion != "" {
		version, err := semver.Parse(source.InitialVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid initial version (%s): %s", source.InitialVersion, err)
		}

		initialVersion = version
	} else {
		initialVersion = semver.Version{Major: 0, Minor: 0, Patch: 0}
	}

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
			InitialVersion: initialVersion,

			Bucket: bucket,
			Key:    source.Key,
		}, nil

	case models.DriverGit:
		return &GitDriver{
			InitialVersion: initialVersion,

			URI:        source.URI,
			Branch:     source.Branch,
			PrivateKey: source.PrivateKey,
			File:       source.File,
			GitUser:    source.GitUser,
		}, nil

	case models.DriverSwift:
		return NewSwiftDriver(&source)

	default:
		return nil, fmt.Errorf("unknown driver: %s", source.Driver)
	}
}
