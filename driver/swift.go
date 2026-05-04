package driver

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
	"github.com/concourse/semver-resource/version"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/objects"
)

type SwiftDriver struct {
	Container          string
	ItemName           string
	InitialVersion     semver.Version
	swiftServiceClient *gophercloud.ServiceClient
}

func NewSwiftDriver(source *models.Source) (Driver, error) {
	os := source.OpenStack
	if os.Container == "" {
		return nil, fmt.Errorf("openstack/container is empty but must be specified")
	}

	if os.Region == "" {
		return nil, fmt.Errorf("openstack/region is empty but must be specified")
	}

	if os.ItemName == "" {
		return nil, fmt.Errorf("openstack/item_name is empty but must be specified")
	}

	opts := gophercloud.AuthOptions{
		IdentityEndpoint:            os.IdentityEndpoint,
		Username:                    os.Username,
		UserID:                      os.UserID,
		Password:                    os.Password,
		DomainID:                    os.DomainID,
		DomainName:                  os.DomainName,
		TenantID:                    os.TenantID,
		TenantName:                  os.TenantName,
		AllowReauth:                 os.AllowReauth,
		TokenID:                     os.TokenID,
		Scope:                       (*gophercloud.AuthScope)(&os.Scope),
		ApplicationCredentialID:     os.ApplicationCredentialID,
		ApplicationCredentialName:   os.ApplicationCredentialName,
		ApplicationCredentialSecret: os.ApplicationCredentialSecret,
	}

	swiftServiceClient, err := getSwiftClient(opts, os.Region)
	if err != nil {
		return nil, err
	}

	_, err = containers.Get(context.TODO(), swiftServiceClient, source.OpenStack.Container, containers.GetOpts{}).ExtractMetadata()
	if err != nil {
		return nil, fmt.Errorf("Unable to get container by name '%s', inner error: %s", source.OpenStack.Container, err.Error())
	}

	var initialVersion semver.Version
	if source.InitialVersion != "" {
		initialVersion, err = semver.Parse(source.InitialVersion)
		if err != nil {
			return nil, fmt.Errorf("Initial version was not a valid sem ver: %s", err.Error())
		}
	} else {
		initialVersion = semver.Version{Major: 0, Minor: 0, Patch: 0}
	}

	driver := &SwiftDriver{
		swiftServiceClient: swiftServiceClient,
		InitialVersion:     initialVersion,
		Container:          source.OpenStack.Container,
		ItemName:           source.OpenStack.ItemName,
	}

	return driver, nil
}

func getSwiftClient(opts gophercloud.AuthOptions, region string) (*gophercloud.ServiceClient, error) {
	provider, err := openstack.AuthenticatedClient(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("Unable to Authenticate, inner error: %s", err.Error())
	}

	swiftServiceClient, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: region,
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to connect object store, inner error: %s", err.Error())
	}

	return swiftServiceClient, nil
}

func createOpts(os models.OpenStackOptions) gophercloud.AuthOptions {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: os.IdentityEndpoint,
		Username:         os.Username,
		UserID:           os.UserID,
		Password:         os.Password,
		DomainID:         os.DomainID,
		DomainName:       os.DomainName,
		TenantID:         os.TenantID,
		TenantName:       os.TenantName,
		AllowReauth:      os.AllowReauth,
		TokenID:          os.TokenID,
	}

	return opts
}

func (driver *SwiftDriver) Bump(bump version.Bump) (semver.Version, error) {
	currentVersion, err := driver.getCurrentVersion()
	if err != nil {
		return semver.Version{}, err
	}

	newVersion := bump.Apply(currentVersion)
	err = driver.Set(newVersion)
	if err != nil {
		return semver.Version{}, err
	}

	return newVersion, nil
}

func (driver *SwiftDriver) Set(newVersion semver.Version) error {
	content := strings.NewReader(newVersion.String())
	opts := objects.CreateOpts{
		Content:            content,
		ContentDisposition: fmt.Sprintf(`attachment; filename="%s"`, driver.ItemName),
	}

	// Now execute the upload
	res := objects.Create(context.TODO(), driver.swiftServiceClient, driver.Container, driver.ItemName, opts)

	// We have the option of extracting the resulting headers from the response
	_, err := res.Extract()
	return err
}

func (driver *SwiftDriver) Check(cursor *semver.Version) ([]semver.Version, error) {
	itemVersion, err := driver.getCurrentVersion()
	if err != nil {
		return nil, err
	}

	return []semver.Version{itemVersion}, nil
}

func (driver *SwiftDriver) getCurrentVersion() (semver.Version, error) {
	downloader := objects.Download(context.TODO(), driver.swiftServiceClient, driver.Container, driver.ItemName, nil)
	bytes, err := downloader.ExtractContent()
	unexpectedResponseCodeError, isType := err.(*gophercloud.ErrUnexpectedResponseCode)
	if isType && unexpectedResponseCodeError.Actual == 404 {
		return driver.InitialVersion, nil
	}

	if err != nil {
		return semver.Version{}, err
	}

	value := strings.TrimSpace(string(bytes))
	itemVersion, err := semver.Parse(value)
	if err != nil {
		return semver.Version{}, fmt.Errorf("parsing number in container: %s", err)
	}

	return itemVersion, nil
}
