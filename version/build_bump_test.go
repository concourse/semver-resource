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

		bump = version.BuildBump{}
	})

	JustBeforeEach(func() {
		outputVersion = bump.Apply(inputVersion)
	})

	Context("when the build is without version number", func() {
		BeforeEach(func() {
			bump.Build = "omega"
			bump.BuildWithoutVersion = true
		})

		Context("when the input is not a build", func() {
			BeforeEach(func() {
				inputVersion.Build = nil
			})

			It("bmps the build without version number", func() {
				Expect(outputVersion).To(Equal(semver.Version{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Build: []string{
						"omega",
					},
				}))
			})
		})

		Context("when the input is a build", func() {
			Context("when the bump is a different build type", func() {
				BeforeEach(func() {
					inputVersion.Build = []string{
						"alpha",
					}
				})

				It("bmps the build without version number", func() {
					Expect(outputVersion).To(Equal(semver.Version{
						Major: 1,
						Minor: 2,
						Patch: 3,
						Build: []string{
							"omega",
						},
					}))
				})
			})

			Context("when the bump is the same build type", func() {
				BeforeEach(func() {
					inputVersion.Build = []string{
						"omega", "1",
					}
				})

				It("bmps the build without version number", func() {
					Expect(outputVersion).To(Equal(semver.Version{
						Major: 1,
						Minor: 2,
						Patch: 3,
						Build: []string{
							"omega",
						},
					}))
				})
			})
		})
	})

	Context("when the version is a build", func() {
		BeforeEach(func() {
			inputVersion.Build = []string{
				"alpha", "1",
			}
		})

		Context("when the bump is the same build type", func() {
			BeforeEach(func() {
				bump.Build = "alpha"
			})

			It("bumps the build version number", func() {
				Expect(outputVersion).To(Equal(semver.Version{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Build: []string{
						"alpha", "2",
					},
				}))
			})

			It("does not mutate the input version", func() {
				Expect(inputVersion).To(Equal(semver.Version{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Build: []string{
						"alpha", "1",
					},
				}))
			})
		})

		Context("when the bump is a different build type", func() {
			BeforeEach(func() {
				bump.Build = "beta"
			})

			It("bumps bumps to version 1 of the new build type", func() {
				Expect(outputVersion).To(Equal(semver.Version{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Build: []string{
						"beta", "1",
					},
				}))
			})
		})
	})

	Context("when the version is not a build", func() {
		BeforeEach(func() {
			inputVersion.Build = nil
		})

		BeforeEach(func() {
			bump.Build = "beta"
		})

		It("bumps bumps to version 1 of the new build type", func() {
			Expect(outputVersion).To(Equal(semver.Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: []string{
					"beta", "1",
				},
			}))
		})
	})
})
