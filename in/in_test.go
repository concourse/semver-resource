package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/concourse/semver-resource/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("In", func() {
	var tmpdir string
	var destination string

	var inCmd *exec.Cmd

	BeforeEach(func() {
		var err error

		tmpdir, err = ioutil.TempDir("", "in-destination")
		Ω(err).ShouldNot(HaveOccurred())

		destination = path.Join(tmpdir, "in-dir")

		inCmd = exec.Command(inPath, destination)
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		var request models.InRequest
		var response models.InResponse

		BeforeEach(func() {
			request = models.InRequest{
				Version: models.Version{},
				Source:  models.Source{},
				Params:  models.Params{},
			}

			response = models.InResponse{}
		})

		JustBeforeEach(func() {
			stdin, err := inCmd.StdinPipe()
			Ω(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(inCmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("with no version", func() {
			for bump, result := range map[string]string{
				"":      "0.0.0",
				"final": "0.0.0",
				"patch": "0.0.1",
				"minor": "0.1.0",
				"major": "1.0.0",
			} {
				bumpLocal := bump
				resultLocal := result

				Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
					BeforeEach(func() {
						request.Params.Bump = bumpLocal
					})

					It("reports "+resultLocal+" as the version", func() {
						Ω(response.Version.Number).Should(Equal(resultLocal))
					})
				})
			}

			Context("when bumping a prerelease", func() {
				BeforeEach(func() {
					request.Params.Pre = "rc"
				})

				for bump, result := range map[string]string{
					"":      "0.0.0-rc.1",
					"final": "0.0.0-rc.1",
					"patch": "0.0.1-rc.1",
					"minor": "0.1.0-rc.1",
					"major": "1.0.0-rc.1",
				} {
					bumpLocal := bump
					resultLocal := result

					Context(fmt.Sprintf("when bumping %s", bumpLocal), func() {
						BeforeEach(func() {
							request.Params.Bump = bumpLocal
						})

						It("reports "+resultLocal+" as the version", func() {
							Ω(response.Version.Number).Should(Equal(resultLocal))
						})
					})
				}
			})
		})

		Context("with a version present", func() {
			BeforeEach(func() {
				request.Version.Number = "1.2.3"
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
						request.Params.Bump = bumpLocal
					})

					It("reports "+resultLocal+" as the version", func() {
						Ω(response.Version.Number).Should(Equal(resultLocal))
					})
				})
			}

			Context("when it's a prerelease", func() {
				BeforeEach(func() {
					request.Version.Number = "1.2.3-rc.1"
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
							request.Params.Bump = bumpLocal
						})

						It("reports "+resultLocal+" as the version", func() {
							Ω(response.Version.Number).Should(Equal(resultLocal))
						})
					})
				}
			})

			Context("when bumping to a prerelease", func() {
				BeforeEach(func() {
					request.Params.Pre = "rc"
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
							request.Params.Bump = bumpLocal
						})

						It("reports "+resultLocal+" as the version", func() {
							Ω(response.Version.Number).Should(Equal(resultLocal))
						})
					})
				}

				Context("when it's already a prerelease", func() {
					BeforeEach(func() {
						request.Version.Number = "1.2.3-rc.1"
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
								request.Params.Bump = bumpLocal
							})

							It("reports "+resultLocal+" as the version", func() {
								Ω(response.Version.Number).Should(Equal(resultLocal))
							})
						})
					}

					Context("of a different type", func() {
						BeforeEach(func() {
							request.Version.Number = "1.2.3-alpha.1"
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
									request.Params.Bump = bumpLocal
								})

								It("reports "+resultLocal+" as the version", func() {
									Ω(response.Version.Number).Should(Equal(resultLocal))
								})
							})
						}
					})
				})
			})
		})
	})
})
