package version_test

import (
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PatchBump", func() {
	var inputVersion semver.Version
	var bump version.Bump
	var outputVersion semver.Version

	BeforeEach(func() {
		inputVersion = semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 3,
			Pre: []semver.PRVersion{
				{VersionStr: "beta"},
				{VersionNum: 1, IsNum: true},
			},
			Build: []string{"b1", "a12b3c4d"},
		}

		bump = version.PatchBump{}
	})

	JustBeforeEach(func() {
		outputVersion = bump.Apply(inputVersion)
	})

	It("bumps patch and zeroes out the subsequent segments while keeping build metadata", func() {
		Expect(outputVersion).To(Equal(semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 4,
			Build: []string{"b1", "a12b3c4d"},
		}))
	})
})
