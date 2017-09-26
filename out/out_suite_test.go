package main_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var outPath string

var accessKeyID = os.Getenv("SEMVER_TESTING_ACCESS_KEY_ID")
var secretAccessKey = os.Getenv("SEMVER_TESTING_SECRET_ACCESS_KEY")
var bucketName = os.Getenv("SEMVER_TESTING_BUCKET")
var regionName = os.Getenv("SEMVER_TESTING_REGION")
var v2signing = os.Getenv("SEMVER_TESTING_V2_SIGNING") == "true"

var _ = BeforeSuite(func() {
	var err error

	Expect(accessKeyID).NotTo(BeEmpty(), "must specify $SEMVER_TESTING_ACCESS_KEY_ID")
	Expect(secretAccessKey).NotTo(BeEmpty(), "must specify $SEMVER_TESTING_SECRET_ACCESS_KEY")
	Expect(bucketName).NotTo(BeEmpty(), "must specify $SEMVER_TESTING_BUCKET")
	Expect(regionName).NotTo(BeEmpty(), "must specify $SEMVER_TESTING_REGION")

	if _, err = os.Stat("/opt/resource/out"); err == nil {
		outPath = "/opt/resource/out"
	} else {
		outPath, err = gexec.Build("github.com/concourse/semver-resource/out")
		Expect(err).NotTo(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestOut(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Out Suite")
}
