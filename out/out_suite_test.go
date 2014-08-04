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

var _ = BeforeSuite(func() {
	var err error

	Ω(accessKeyID).ShouldNot(BeEmpty(), "must specify $SEMVER_TESTING_ACCESS_KEY_ID")
	Ω(secretAccessKey).ShouldNot(BeEmpty(), "must specify $SEMVER_TESTING_SECRET_ACCESS_KEY")
	Ω(bucketName).ShouldNot(BeEmpty(), "must specify $SEMVER_TESTING_BUCKET")

	outPath, err = gexec.Build("github.com/concourse/semver-resource/out")
	Ω(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestOut(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Out Suite")
}
