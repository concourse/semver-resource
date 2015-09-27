package version_test

import (
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PreBump", func() {
	var inputVersion semver.Version
	var bump version.PreBump
	var outputVersion semver.Version

	BeforeEach(func() {
		inputVersion = semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 3,
		}

		bump = version.PreBump{}
	})

	JustBeforeEach(func() {
		outputVersion = bump.Apply(inputVersion)
	})

	Context("when the version is a prerelease", func() {
		BeforeEach(func() {
			inputVersion.Pre = []semver.PRVersion{
				{VersionStr: "alpha"},
				{VersionNum: 1, IsNum: true},
			}
		})

		Context("when the bump is the same prerelease type", func() {
			BeforeEach(func() {
				bump.Pre = "alpha"
			})

			It("bumps the prerelease version number", func() {
				Expect(outputVersion).To(Equal(semver.Version{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Pre: []semver.PRVersion{
						{VersionStr: "alpha"},
						{VersionNum: 2, IsNum: true},
					},
				}))
			})

			It("does not mutate the input version", func() {
				Expect(inputVersion).To(Equal(semver.Version{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Pre: []semver.PRVersion{
						{VersionStr: "alpha"},
						{VersionNum: 1, IsNum: true},
					},
				}))
			})
		})

		Context("when the bump is a different prerelease type", func() {
			BeforeEach(func() {
				bump.Pre = "beta"
			})

			It("bumps bumps to version 1 of the new prerelease type", func() {
				Expect(outputVersion).To(Equal(semver.Version{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Pre: []semver.PRVersion{
						{VersionStr: "beta"},
						{VersionNum: 1, IsNum: true},
					},
				}))
			})
		})
	})

	Context("when the version is not a prerelease", func() {
		BeforeEach(func() {
			inputVersion.Pre = nil
		})

		BeforeEach(func() {
			bump.Pre = "beta"
		})

		It("bumps bumps to version 1 of the new prerelease type", func() {
			Expect(outputVersion).To(Equal(semver.Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Pre: []semver.PRVersion{
					{VersionStr: "beta"},
					{VersionNum: 1, IsNum: true},
				},
			}))
		})
	})
})
