package version_test

import (
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildBump", func() {
	var inputVersion semver.Version
	var bump version.BuildBump
	var outputVersion semver.Version

	BeforeEach(func() {
		inputVersion = semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 3,
		}

		bump = version.BuildBump{[]string{"b1","a12b3c4d"}}
	})

	JustBeforeEach(func() {
		outputVersion = bump.Apply(inputVersion)
	})

	It("appends the build metadata to the end of the version", func() {
		Expect(outputVersion).To(Equal(semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 3,
			Build: []string{"b1", "a12b3c4d"},
		}))
	})
})
