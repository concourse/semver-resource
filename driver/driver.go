package driver

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
	"github.com/concourse/semver-resource/version"
)

type Driver interface {
	Bump(version.Bump) (semver.Version, error)
	Set(semver.Version) error
	Check(*semver.Version) ([]semver.Version, error)
}

const maxRetries = 12

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
		var creds *credentials.Credentials

		if source.AccessKeyID == "" && source.SecretAccessKey == "" {
			creds = credentials.AnonymousCredentials
		} else {
			creds = credentials.NewStaticCredentials(source.AccessKeyID, source.SecretAccessKey, "")
		}

		regionName := source.RegionName
		if len(regionName) == 0 {
			regionName = "us-east-1"
		}

		awsConfig := &aws.Config{
			Region:           aws.String(regionName),
			Credentials:      creds,
			S3ForcePathStyle: aws.Bool(true),
			MaxRetries:       aws.Int(maxRetries),
			DisableSSL:       aws.Bool(source.DisableSSL),
		}

		if len(source.Endpoint) != 0 {
			awsConfig.Endpoint = aws.String(source.Endpoint)
		}

		svc := s3.New(session.New(awsConfig))

		return &S3Driver{
			InitialVersion: initialVersion,

			Svc:                  svc,
			BucketName:           source.Bucket,
			Key:                  source.Key,
			ServerSideEncryption: source.ServerSideEncryption,
		}, nil

	case models.DriverGit:
		return &GitDriver{
			InitialVersion: initialVersion,

			URI:        source.URI,
			Branch:     source.Branch,
			PrivateKey: source.PrivateKey,
			Username:   source.Username,
			Password:   source.Password,
			File:       source.File,
			GitUser:    source.GitUser,
		}, nil

	case models.DriverSwift:
		return NewSwiftDriver(&source)

	default:
		return nil, fmt.Errorf("unknown driver: %s", source.Driver)
	}
}
