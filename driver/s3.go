package driver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
)

type S3Driver struct {
	InitialVersion semver.Version

	Svc        *s3.S3
	BucketName string
	Key        string
}

func (driver *S3Driver) Bump(bump version.Bump) (semver.Version, error) {
	var currentVersion semver.Version

	resp, err := driver.Svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(driver.BucketName),
		Key:    aws.String(driver.Key),
	})
	if err == nil {
		bucketNumberPayload, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return semver.Version{}, err
		}
		defer resp.Body.Close()

		payloadStr := strings.TrimSpace(string(bucketNumberPayload))
		currentVersion, err = semver.Parse(payloadStr)
		if err != nil {
			return semver.Version{}, err
		}
	} else if s3err, ok := err.(awserr.RequestFailure); ok && s3err.StatusCode() == 404 {
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
	params := &s3.PutObjectInput{
		Bucket:      aws.String(driver.BucketName),
		Key:         aws.String(driver.Key),
		ContentType: aws.String("text/plain"),
		Body:        bytes.NewReader([]byte(newVersion.String())),
		ACL:         aws.String(s3.ObjectCannedACLPrivate),
	}

	_, err := driver.Svc.PutObject(params)
	return err
}

func (driver *S3Driver) Check(cursor *semver.Version) ([]semver.Version, error) {
	var bucketNumber string

	resp, err := driver.Svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(driver.BucketName),
		Key:    aws.String(driver.Key),
	})
	if err == nil {
		bucketNumberPayload, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []semver.Version{}, err
		}
		defer resp.Body.Close()

		bucketNumber = string(bucketNumberPayload)
	} else if s3err, ok := err.(awserr.RequestFailure); ok && s3err.StatusCode() == 404 {
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

	if cursor == nil || bucketVersion.GTE(*cursor) {
		return []semver.Version{bucketVersion}, nil
	}

	return []semver.Version{}, nil
}
