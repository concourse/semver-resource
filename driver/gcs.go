package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"github.com/blang/semver"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"

	"github.com/concourse/semver-resource/version"
)

type GCSDriver struct {
	InitialVersion semver.Version

	Servicer   IOServicer
	BucketName string
	Key        string
}

func (d *GCSDriver) Bump(b version.Bump) (semver.Version, error) {
	versions, err := d.Check(nil)

	if err != nil {
		return semver.Version{}, err
	}

	if len(versions) == 0 {
		return semver.Version{}, nil
	}

	newVersion := b.Apply(versions[0])
	err = d.Set(newVersion)

	if err != nil {
		return semver.Version{}, err
	}
	return newVersion, nil
}

func (d *GCSDriver) Set(v semver.Version) error {
	w, err := d.Servicer.PutObject(d.BucketName, d.Key)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(v.String()))
	if err != nil {
		return err
	}
	return w.Close()
}

func (d *GCSDriver) Check(cursor *semver.Version) ([]semver.Version, error) {
	r, err := d.Servicer.GetObject(d.BucketName, d.Key)

	if errors.Is(err, storage.ErrObjectNotExist) {
		if cursor == nil {
			return []semver.Version{d.InitialVersion}, nil
		}
		return []semver.Version{}, nil
	} else if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	v, err := semver.Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("parsing number in bucket: %s", err)
	}

	return []semver.Version{v}, nil
}

type IOServicer interface {
	GetObject(bucketName, objectName string) (io.ReadCloser, error)
	PutObject(bucketName, objectName string) (io.WriteCloser, error)
}

type GCSIOServicer struct {
	JSONCredentials string
	Token           string
}

func (s *GCSIOServicer) authOption() (option.ClientOption, error) {
	if s.Token != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: s.Token,
		})
		return option.WithTokenSource(tokenSource), nil
	}

	temp, err := os.CreateTemp("", "auth-credentials.json")
	if err != nil {
		return nil, err
	}

	_, err = temp.WriteString(s.JSONCredentials)
	if err != nil {
		return nil, err
	}
	// Close the file so the credentials can be read by the client
	temp.Close()

	return option.WithCredentialsFile(temp.Name()), nil
}

func (s *GCSIOServicer) GetObject(bucketName, objectName string) (io.ReadCloser, error) {
	authOpt, err := s.authOption()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, authOpt)
	if err != nil {
		return nil, err
	}

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(objectName)

	return obj.NewReader(context.Background())
}

func (s *GCSIOServicer) PutObject(bucketName, objectName string) (io.WriteCloser, error) {
	authOpt, err := s.authOption()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, authOpt)
	if err != nil {
		return nil, err
	}

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(objectName)

	w := obj.NewWriter(context.Background())
	w.CacheControl = "private"
	return w, nil
}
