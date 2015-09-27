package version_test

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
	. "github.com/concourse/semver-resource/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BumpParams", func() {
	var (
		version *semver.Version
		params  models.InParams
	)

	BeforeEach(func() {
		version = &semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 3,
		}

		params = models.InParams{}
	})

	JustBeforeEach(func() {
		BumpParams(version, params)
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
				params.Bump = bumpLocal
			})

			It("bumps to "+resultLocal, func() {
				Ω(version.String()).Should(Equal(resultLocal))
			})
		})
	}

	Context("when bumping to a prerelease", func() {
		BeforeEach(func() {
			params.Pre = "rc"
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
					params.Bump = bumpLocal
				})

				It("bumps to "+resultLocal, func() {
					Ω(version.String()).Should(Equal(resultLocal))
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
				"final": "1.2.3-rc.2",
				"patch": "1.2.3-rc.2",
				"minor": "1.2.3-rc.2",
				"major": "1.2.3-rc.2",
			} {
				bumpLocal := bump
				resultLocal := result

				Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
					BeforeEach(func() {
						params.Bump = bumpLocal
					})

					It("bumps to "+resultLocal, func() {
						Ω(version.String()).Should(Equal(resultLocal))
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
					"patch": "1.2.3-rc.1",
					"minor": "1.2.3-rc.1",
					"major": "1.2.3-rc.1",
				} {
					bumpLocal := bump
					resultLocal := result

					Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
						BeforeEach(func() {
							params.Bump = bumpLocal
						})

						It("bumps to "+resultLocal, func() {
							Ω(version.String()).Should(Equal(resultLocal))
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
					params.Bump = bumpLocal
				})

				It("bumps to "+resultLocal, func() {
					Ω(version.String()).Should(Equal(resultLocal))
				})
			})
		}
	})
})
