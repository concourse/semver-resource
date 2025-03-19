package driver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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
		var credsProvider aws.CredentialsProvider

		if source.AccessKeyID != "" && source.SecretAccessKey != "" {
			credsProvider = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(source.AccessKeyID, source.SecretAccessKey, source.SessionToken))
			_, err := credsProvider.Retrieve(context.Background())
			if err != nil {
				return nil, err
			}
		}

		regionName := source.RegionName
		if regionName == "" {
			regionName = "us-east-1"
		}

		var httpClient *http.Client
		if source.SkipSSLVerification {
			httpClient = &http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}}
		} else {
			httpClient = http.DefaultClient
		}

		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(regionName),
			config.WithHTTPClient(httpClient),
			config.WithRetryMaxAttempts(maxRetries),
			config.WithCredentialsProvider(credsProvider),
		)
		if err != nil {
			return nil, fmt.Errorf("error loading default aws config: %w", err)
		}

		if source.AssumeRoleArn != "" {
			stsClient := sts.NewFromConfig(cfg)
			roleCreds := stscreds.NewAssumeRoleProvider(stsClient, source.AssumeRoleArn)
			creds, err := roleCreds.Retrieve(context.TODO())
			if err != nil {
				return nil, fmt.Errorf("error assuming role: %w", err)
			}

			cfg.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
				creds.AccessKeyID,
				creds.SecretAccessKey,
				creds.SessionToken,
			))
		}

		s3Opts := []func(*s3.Options){
			func(o *s3.Options) {
				o.UsePathStyle = true
			},
		}

		if source.Endpoint != "" {
			endpoint := source.Endpoint
			u, err := url.Parse(source.Endpoint)
			if err != nil {
				return nil, fmt.Errorf("error parsing given endpoint: %w", err)
			}
			if u.Scheme == "" {
				// source.Endpoint is a hostname
				scheme := "https://"
				if source.DisableSSL {
					scheme = "http://"
				}
				endpoint = scheme + source.Endpoint
			}

			s3Opts = append(s3Opts, func(o *s3.Options) {
				o.BaseEndpoint = &endpoint
			})
		}

		s3Client := s3.NewFromConfig(cfg, s3Opts...)

		return &S3Driver{
			InitialVersion: initialVersion,

			Svc:                  s3Client,
			BucketName:           source.Bucket,
			Key:                  source.Key,
			ServerSideEncryption: source.ServerSideEncryption,
		}, nil

	case models.DriverGit:
		return &GitDriver{
			InitialVersion: initialVersion,

			URI:                 source.URI,
			Branch:              source.Branch,
			PrivateKey:          source.PrivateKey,
			Username:            source.Username,
			Password:            source.Password,
			File:                source.File,
			GitUser:             source.GitUser,
			CommitMessage:       source.CommitMessage,
			SkipSSLVerification: source.SkipSSLVerification,
		}, nil

	case models.DriverSwift:
		return NewSwiftDriver(&source)

	case models.DriverGCS:
		servicer := &GCSIOServicer{
			JSONCredentials: source.JSONKey,
		}

		return &GCSDriver{
			InitialVersion: initialVersion,

			Servicer:   servicer,
			BucketName: source.Bucket,
			Key:        source.Key,
		}, nil

	default:
		return nil, fmt.Errorf("unknown driver: %s", source.Driver)
	}
}
