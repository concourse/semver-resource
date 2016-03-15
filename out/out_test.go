package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/concourse/semver-resource/models"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Out", func() {
	var key string

	var source string

	var outCmd *exec.Cmd

	BeforeEach(func() {
		var err error

		source, err = ioutil.TempDir("", "out-source")
		Expect(err).NotTo(HaveOccurred())

		outCmd = exec.Command(outPath, source)
	})

	AfterEach(func() {
		os.RemoveAll(source)
	})

	Context("when executed", func() {
		var request models.OutRequest
		var response models.OutResponse

		var bucket *s3.Bucket

		BeforeEach(func() {
			guid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())

			key = guid.String()

			auth := aws.Auth{
				AccessKey: accessKeyID,
				SecretKey: secretAccessKey,
			}

			region, ok := aws.Regions[regionName]
			Expect(ok).To(BeTrue())

			client := s3.New(auth, region)

			bucket = client.Bucket(bucketName)

			request = models.OutRequest{
				Version: models.Version{},
				Source: models.Source{
					Bucket:          bucketName,
					Key:             key,
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
					RegionName:      regionName,
				},
				Params: models.OutParams{},
			}

			response = models.OutResponse{}
		})

		AfterEach(func() {
			err := bucket.Del(key)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			stdin, err := outCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(outCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			// account for roundtrip to s3
			Eventually(session, 5*time.Second).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when setting the version", func() {
			BeforeEach(func() {
				request.Params.File = "number"
			})

			Context("when a valid version is in the file", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(filepath.Join(source, "number"), []byte("1.2.3"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("reports the version as the resource's version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					contents, err := bucket.Get(key)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(contents)).To(Equal("1.2.3"))
				})
			})
		})

		Context("when bumping the version", func() {
			BeforeEach(func() {
				err := bucket.Put(key, []byte("1.2.3"), "text/plain", s3.Private)
				Expect(err).NotTo(HaveOccurred())
			})

			for bump, result := range map[string]string{
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

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal(resultLocal))
					})

					It("saves the contents of the file in the configured bucket", func() {
						contents, err := bucket.Get(key)
						Expect(err).NotTo(HaveOccurred())

						Expect(string(contents)).To(Equal(resultLocal))
					})
				})
			}
		})

		Context("when bumping the version to a prerelease", func() {
			BeforeEach(func() {
				request.Params.Pre = "alpha"
			})

			Context("when the version is not a prerelease", func() {
				BeforeEach(func() {
					err := bucket.Put(key, []byte("1.2.3"), "text/plain", s3.Private)
					Expect(err).NotTo(HaveOccurred())
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.1"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					contents, err := bucket.Get(key)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(contents)).To(Equal("1.2.3-alpha.1"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0-alpha.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						contents, err := bucket.Get(key)
						Expect(err).NotTo(HaveOccurred())

						Expect(string(contents)).To(Equal("1.3.0-alpha.1"))
					})
				})
			})

			Context("when the version is the same prerelease", func() {
				BeforeEach(func() {
					err := bucket.Put(key, []byte("1.2.3-alpha.2"), "text/plain", s3.Private)
					Expect(err).NotTo(HaveOccurred())
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.3"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					contents, err := bucket.Get(key)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(contents)).To(Equal("1.2.3-alpha.3"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0-alpha.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						contents, err := bucket.Get(key)
						Expect(err).NotTo(HaveOccurred())

						Expect(string(contents)).To(Equal("1.3.0-alpha.1"))
					})
				})
			})

			Context("when the version is a different prerelease", func() {
				BeforeEach(func() {
					err := bucket.Put(key, []byte("1.2.3-beta.2"), "text/plain", s3.Private)
					Expect(err).NotTo(HaveOccurred())
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.1"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					contents, err := bucket.Get(key)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(contents)).To(Equal("1.2.3-alpha.1"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0-alpha.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						contents, err := bucket.Get(key)
						Expect(err).NotTo(HaveOccurred())

						Expect(string(contents)).To(Equal("1.3.0-alpha.1"))
					})
				})
			})
		})
	})
})
