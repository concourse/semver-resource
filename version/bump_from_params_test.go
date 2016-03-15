package version_test

import (
	"fmt"

	"github.com/blang/semver"
	. "github.com/concourse/semver-resource/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BumpForParams", func() {
	var (
		version semver.Version

		bumpParam string
		preParam  string
	)

	BeforeEach(func() {
		version = semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 3,
		}

		bumpParam = ""
		preParam = ""
	})

	JustBeforeEach(func() {
		version = BumpFromParams(bumpParam, preParam).Apply(version)
	})

	for bump, result := range map[string]string{
		"":      "1.2.3",
		"final": "1.2.3",
		"patch": "1.2.4",
		"minor": "1.3.0",
		"major": "2.0.0",
	} {
		bumpLocal := bump
		resultLocal := result

		Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
			BeforeEach(func() {
				bumpParam = bumpLocal
			})

			It("bumps to "+resultLocal, func() {
				Expect(version.String()).To(Equal(resultLocal))
			})
		})
	}

	Context("when bumping to a prerelease", func() {
		BeforeEach(func() {
			preParam = "rc"
		})

		for bump, result := range map[string]string{
			"":      "1.2.3-rc.1",
			"final": "1.2.3-rc.1",
			"patch": "1.2.4-rc.1",
			"minor": "1.3.0-rc.1",
			"major": "2.0.0-rc.1",
		} {
			bumpLocal := bump
			resultLocal := result

			Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
				BeforeEach(func() {
					bumpParam = bumpLocal
				})

				It("bumps to "+resultLocal, func() {
					Expect(version.String()).To(Equal(resultLocal))
				})
			})
		}

		Context("when it's already a prerelease", func() {
			BeforeEach(func() {
				version.Pre = []semver.PRVersion{
					{VersionStr: "rc"},
					{VersionNum: 1, IsNum: true},
				}
			})

			for bump, result := range map[string]string{
				"":      "1.2.3-rc.2",
				"final": "1.2.3-rc.1",
				"patch": "1.2.4-rc.1",
				"minor": "1.3.0-rc.1",
				"major": "2.0.0-rc.1",
			} {
				bumpLocal := bump
				resultLocal := result

				Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
					BeforeEach(func() {
						bumpParam = bumpLocal
					})

					It("bumps to "+resultLocal, func() {
						Expect(version.String()).To(Equal(resultLocal))
					})
				})
			}

			Context("of a different type", func() {
				BeforeEach(func() {
					version.Pre[0].VersionStr = "different-type"
				})

				for bump, result := range map[string]string{
					"":      "1.2.3-rc.1",
					"final": "1.2.3-rc.1",
					"patch": "1.2.4-rc.1",
					"minor": "1.3.0-rc.1",
					"major": "2.0.0-rc.1",
				} {
					bumpLocal := bump
					resultLocal := result

					Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
						BeforeEach(func() {
							bumpParam = bumpLocal
						})

						It("bumps to "+resultLocal, func() {
							Expect(version.String()).To(Equal(resultLocal))
						})
					})
				}
			})
		})
	})

	Context("when bumping from a prerelease", func() {
		BeforeEach(func() {
			version.Pre = []semver.PRVersion{
				{VersionStr: "rc"},
				{VersionNum: 1, IsNum: true},
			}
		})

		for bump, result := range map[string]string{
			"":      "1.2.3-rc.1",
			"final": "1.2.3",
			"patch": "1.2.4",
			"minor": "1.3.0",
			"major": "2.0.0",
		} {
			bumpLocal := bump
			resultLocal := result

			Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
				BeforeEach(func() {
					bumpParam = bumpLocal
				})

				It("bumps to "+resultLocal, func() {
					Expect(version.String()).To(Equal(resultLocal))
				})
			})
		}
	})
})
