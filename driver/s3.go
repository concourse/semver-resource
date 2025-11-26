package driver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
)

type Servicer interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3Driver struct {
	InitialVersion semver.Version

	Svc                  Servicer
	BucketName           string
	Key                  string
	ServerSideEncryption string
	ChecksumAlgorithm    types.ChecksumAlgorithm
}

func (driver *S3Driver) Bump(bump version.Bump) (semver.Version, error) {
	var currentVersion semver.Version

	resp, err := driver.Svc.GetObject(context.TODO(),
		&s3.GetObjectInput{
			Bucket: aws.String(driver.BucketName),
			Key:    aws.String(driver.Key),
		})
	if err == nil {
		bucketNumberPayload, err := io.ReadAll(resp.Body)
		if err != nil {
			return semver.Version{}, err
		}
		defer resp.Body.Close()

		payloadStr := strings.TrimSpace(string(bucketNumberPayload))
		currentVersion, err = semver.Parse(payloadStr)
		if err != nil {
			return semver.Version{}, err
		}
	} else {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			currentVersion = driver.InitialVersion
		} else {
			return semver.Version{}, err
		}
	}

	newVersion := bump.Apply(currentVersion)

	err = driver.Set(newVersion)
	if err != nil {
		return semver.Version{}, err
	}

	return newVersion, nil
}

func (driver *S3Driver) Set(newVersion semver.Version) error {
	params := &s3.PutObjectInput{
		Bucket:      aws.String(driver.BucketName),
		Key:         aws.String(driver.Key),
		ContentType: aws.String("text/plain"),
		Body:        bytes.NewReader([]byte(newVersion.String())),
	}

	if len(driver.ServerSideEncryption) > 0 {
		params.ServerSideEncryption = types.ServerSideEncryption(driver.ServerSideEncryption)
	}

	if len(driver.ChecksumAlgorithm) > 0 {
		params.ChecksumAlgorithm = driver.ChecksumAlgorithm
	}

	_, err := driver.Svc.PutObject(context.TODO(), params)
	return err
}

func (driver *S3Driver) Check(cursor *semver.Version) ([]semver.Version, error) {
	var bucketNumber string

	resp, err := driver.Svc.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(driver.BucketName),
		Key:    aws.String(driver.Key),
	})
	if err == nil {
		bucketNumberPayload, err := io.ReadAll(resp.Body)
		if err != nil {
			return []semver.Version{}, err
		}
		defer resp.Body.Close()

		bucketNumber = string(bucketNumberPayload)
	} else {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			if cursor == nil {
				return []semver.Version{driver.InitialVersion}, nil
			} else {
				return []semver.Version{}, nil
			}
		} else {
			return nil, err
		}
	}

	bucketVersion, err := semver.Parse(bucketNumber)
	if err != nil {
		return nil, fmt.Errorf("parsing number in bucket: %s", err)
	}

	return []semver.Version{bucketVersion}, nil
}
