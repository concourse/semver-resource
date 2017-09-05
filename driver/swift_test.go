package driver

import (
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
	"github.com/concourse/semver-resource/version"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/rackspace/gophercloud/openstack/objectstorage/v1/objects"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var client *gophercloud.ServiceClient
var containerName = fmt.Sprintf("test_container_%d", config.GinkgoConfig.ParallelNode)

var _ = Describe("Swift", func() {
	BeforeSuite(func() {
		identityEndpoint := os.Getenv("OS_AUTH_URL")
		tenantID := os.Getenv("OS_TENANT_ID")
		tenantName := os.Getenv("OS_TENANT_NAME")
		username := os.Getenv("OS_USERNAME")
		password := os.Getenv("OS_PASSWORD")
		region := os.Getenv("OS_REGION_NAME")

		if identityEndpoint != "" || tenantID != "" || tenantName != "" || username != "" || password != "" || region != "" {
			opts := gophercloud.AuthOptions{
				IdentityEndpoint: identityEndpoint,
				Username:         username,
				Password:         password,
				TenantID:         tenantID,
				TenantName:       tenantName,
			}
			var err error
			client, err = getSwiftClient(opts, region)
			Expect(err).To(BeNil())

			err = createContainer(containerName)
			Expect(err).To(BeNil())
		}
	})

	AfterSuite(func() {
		if client != nil {
			err := deleteContainer(containerName)
			Expect(err).To(BeNil())
		}
	})

	BeforeEach(func() {
		if client == nil {
			Skip("env vars not set, skipping Swift tests")
		}
	})

	It("NewSwiftDriver with empty container name should fail", func() {
		driver, err := NewSwiftDriver(
			&models.Source{
				OpenStack: models.OpenStackOptions{
					Region: "region", ItemName: "itemName",
				},
			})

		Expect(driver).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("openstack/container is empty but must be specified"))
	})

	It("NewSwiftDriver with empty item_name name should fail", func() {
		driver, err := NewSwiftDriver(
			&models.Source{
				OpenStack: models.OpenStackOptions{
					Region: "region", Container: "c",
				},
			})

		Expect(driver).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("openstack/item_name is empty but must be specified"))
	})

	It("NewSwiftDriver with empty region name should fail", func() {
		driver, err := NewSwiftDriver(
			&models.Source{
				OpenStack: models.OpenStackOptions{
					ItemName: "i", Container: "c",
				},
			})

		Expect(driver).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("openstack/region is empty but must be specified"))
	})

	It("returns the the initial version as no source is present", func() {
		driver, err := newTestSwiftDriver("1.0.0", "testitem1.txt")
		defer deleteObject("testitem1.txt")
		Expect(err).To(BeNil())

		semVers, err := driver.Check(nil)
		Expect(err).To(BeNil())
		Expect(semVers).To(HaveLen(1))
		Expect(semVers[0].String()).Should(Equal("1.0.0"))
	})

	It("bumps the initial when no source is present", func() {
		driver, err := newTestSwiftDriver("1.0.0", "testitem2.txt")
		defer deleteObject("testitem2.txt")
		Expect(err).To(BeNil())
		semVer, err := driver.Bump(version.PatchBump{})
		Expect(err).To(BeNil())
		Expect(semVer.String()).To(Equal("1.0.1"))
	})

	It("bump should return increased semver than one in objectstore", func() {
		driver, err := newTestSwiftDriver("1.0.0", "testitem3.txt")
		defer deleteObject("testitem3.txt")
		Expect(err).To(BeNil())
		// Setup test with version in object store
		err = driver.Set(semver.Version{Major: 2, Minor: 0, Patch: 10})
		Expect(err).To(BeNil())

		semVer, err := driver.Bump(version.PatchBump{})
		Expect(err).To(BeNil())
		Expect(semVer.String()).To(Equal("2.0.11"))
	})

	It("check returns empty if current in swift is less than supplied version", func() {
		driver, err := newTestSwiftDriver("1.0.0", "testitem3.txt")
		defer deleteObject("testitem3.txt")
		Expect(err).To(BeNil())
		err = driver.Set(semver.Version{Major: 1, Minor: 0, Patch: 10})
		Expect(err).To(BeNil())

		greaterThanVersion := semver.Version{Major: 2, Minor: 0, Patch: 0}
		semVers, err := driver.Check(&greaterThanVersion)
		Expect(err).To(BeNil())
		Expect(semVers).To(BeEmpty())
	})

	It("check returns current version if it is the same as supplied version", func() {
		driver, err := newTestSwiftDriver("1.0.0", "testitem3.txt")
		defer deleteObject("testitem3.txt")
		Expect(err).Should(BeNil())
		err = driver.Set(semver.Version{Major: 2, Minor: 0, Patch: 10})
		Expect(err).Should(BeNil())

		sameVersion := semver.Version{Major: 2, Minor: 0, Patch: 10}
		semVers, err := driver.Check(&sameVersion)
		Expect(err).To(BeNil())
		Expect(semVers).To(HaveLen(1))
		Expect(semVers[0].String()).To(Equal("2.0.10"))
	})

	It("check should return current semver if greater than supplied version", func() {
		driver, err := newTestSwiftDriver("1.0.0", "testitem3.txt")
		defer deleteObject("testitem3.txt")
		Expect(err).Should(BeNil())
		err = driver.Set(semver.Version{Major: 2, Minor: 0, Patch: 10})
		Expect(err).Should(BeNil())

		lessThanVersion := semver.Version{Major: 1, Minor: 0, Patch: 0}
		semVers, err := driver.Check(&lessThanVersion)
		Expect(err).To(BeNil())
		Expect(semVers).To(HaveLen(1))
		Expect(semVers[0].String()).To(Equal("2.0.10"))
	})

	It("check should return current semver if greater supplied version is nil", func() {
		driver, err := newTestSwiftDriver("1.0.0", "testitem3.txt")
		defer deleteObject("testitem3.txt")
		Expect(err).To(BeNil())
		err = driver.Set(semver.Version{Major: 2, Minor: 0, Patch: 10})
		Expect(err).To(BeNil())

		semVers, err := driver.Check(nil)
		Expect(err).To(BeNil())
		Expect(semVers).To(HaveLen(1))
		Expect(semVers[0].String()).To(Equal("2.0.10"))
	})
})

func newTestSwiftDriver(initialVersion string, itemName string) (Driver, error) {

	identityEndpoint := os.Getenv("OS_AUTH_URL")
	tenantID := os.Getenv("OS_TENANT_ID")
	tenantName := os.Getenv("OS_TENANT_NAME")
	username := os.Getenv("OS_USERNAME")
	password := os.Getenv("OS_PASSWORD")
	region := os.Getenv("OS_REGION_NAME")

	os := models.OpenStackOptions{
		IdentityEndpoint: identityEndpoint,
		Username:         username,
		Password:         password,
		TenantID:         tenantID,
		TenantName:       tenantName,
		Region:           region,
		Container:        containerName,
	}

	os.ItemName = itemName
	source := models.Source{OpenStack: os, InitialVersion: initialVersion}
	return NewSwiftDriver(&source)
}

func createContainer(containerName string) error {
	if client == nil {
		return fmt.Errorf("Can't create a container because swift client is nil")
	}

	res := containers.Create(client, containerName, nil)

	_, err := res.Extract()
	return err
}

func deleteContainer(containerName string) error {
	if client == nil {
		return fmt.Errorf("Can't delete a container because swift client is nil")
	}

	res := containers.Delete(client, containerName)

	_, err := res.Extract()
	return err
}

func deleteObject(objectName string) error {
	if client == nil {
		return fmt.Errorf("Can't delete an object because swift client is nil")
	}

	res := objects.Delete(client, containerName, objectName, nil)

	_, err := res.Extract()
	return err
}
