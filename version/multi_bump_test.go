package version_test

import (
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MultiBump", func() {
	var inputVersion semver.Version
	var bump version.MultiBump
	var outputVersion semver.Version

	BeforeEach(func() {
		inputVersion = semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 3,
		}

		bump = version.MultiBump{
			version.MajorBump{},
			version.MinorBump{},
			version.PatchBump{},
			version.PatchBump{},
			version.PatchBump{},
			version.PreBump{"beta", true},
			version.PreBump{"beta", false},
		}
	})

	JustBeforeEach(func() {
		outputVersion = bump.Apply(inputVersion)
	})

	It("applies the bumps in order", func() {
		Expect(outputVersion).To(Equal(semver.Version{
			Major: 2,
			Minor: 1,
			Patch: 3,
			Pre: []semver.PRVersion{
				{VersionStr: "beta"},
				{VersionNum: 1, IsNum: true},
			},
		}))
	})
})
