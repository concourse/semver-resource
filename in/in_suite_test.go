package main_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var inPath string

var accessKeyID = os.Getenv("SEMVER_TESTING_ACCESS_KEY_ID")
var secretAccessKey = os.Getenv("SEMVER_TESTING_SECRET_ACCESS_KEY")
var bucketName = os.Getenv("SEMVER_TESTING_BUCKET")
var regionName = os.Getenv("SEMVER_TESTING_REGION")

var _ = BeforeSuite(func() {
	var err error

	Ω(accessKeyID).ShouldNot(BeEmpty(), "must specify $SEMVER_TESTING_ACCESS_KEY_ID")
	Ω(secretAccessKey).ShouldNot(BeEmpty(), "must specify $SEMVER_TESTING_SECRET_ACCESS_KEY")
	Ω(bucketName).ShouldNot(BeEmpty(), "must specify $SEMVER_TESTING_BUCKET")
	Ω(regionName).ShouldNot(BeEmpty(), "must specify $SEMVER_TESTING_REGION")

	inPath, err = gexec.Build("github.com/concourse/semver-resource/in")
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestIn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "In Suite")
}
