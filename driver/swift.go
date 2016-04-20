package driver

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
	"github.com/concourse/semver-resource/version"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/rackspace/gophercloud/openstack/objectstorage/v1/objects"
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
		IdentityEndpoint: os.IdentityEndpoint,
		Username:         os.Username,
		UserID:           os.UserID,
		Password:         os.Password,
		APIKey:           os.APIKey,
		DomainID:         os.DomainID,
		DomainName:       os.DomainName,
		TenantID:         os.TenantID,
		TenantName:       os.TenantName,
		AllowReauth:      os.AllowReauth,
		TokenID:          os.TokenID,
	}

	swiftServiceClient, err := getSwiftClient(opts, os.Region)
	if err != nil {
		return nil, err
	}

	_, err = containers.Get(swiftServiceClient, source.OpenStack.Container).ExtractMetadata()
	if err != nil {
		return nil, fmt.Errorf("Unable to get container by name '%s'", source.OpenStack.Container)
	}

	initialVersion, err := semver.Parse(source.InitialVersion)
	if err != nil {
		return nil, fmt.Errorf("Initial version was not a valid sem ver: %s", err.Error())
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
	provider, err := openstack.AuthenticatedClient(opts)
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
		APIKey:           os.APIKey,
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
		ContentDisposition: fmt.Sprintf(`attachment; filename="%s"`, driver.ItemName),
	}

	// Now execute the upload
	res := objects.Create(driver.swiftServiceClient, driver.Container, driver.ItemName, content, opts)

	// We have the option of extracting the resulting headers from the response
	_, err := res.ExtractHeader()
	return err
}

func (driver *SwiftDriver) Check(cursor *semver.Version) ([]semver.Version, error) {
	itemVersion, err := driver.getCurrentVersion()
	if err != nil {
		return nil, err
	}

	if cursor == nil || itemVersion.GTE(*cursor) {
		return []semver.Version{itemVersion}, nil
	}

	return []semver.Version{}, nil
}

func (driver *SwiftDriver) getCurrentVersion() (semver.Version, error) {
	bytes, err := objects.Download(driver.swiftServiceClient, driver.Container, driver.ItemName, nil).ExtractContent()
	unexpectedResponseCodeError, isType := err.(*gophercloud.UnexpectedResponseCodeError)
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
