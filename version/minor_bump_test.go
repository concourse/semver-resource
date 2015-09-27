package version_test

import (
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MinorBump", func() {
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
		}

		bump = version.MinorBump{}
	})

	JustBeforeEach(func() {
		outputVersion = bump.Apply(inputVersion)
	})

	It("bumps minor and zeroes out the subsequent segments", func() {
		Expect(outputVersion).To(Equal(semver.Version{
			Major: 1,
			Minor: 3,
			Patch: 0,
		}))
	})
})
