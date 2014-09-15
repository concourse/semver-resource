package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/concourse/semver-resource/models"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Check", func() {
	var key string

	var tmpdir string
	var destination string

	var checkCmd *exec.Cmd

	BeforeEach(func() {
		var err error

		tmpdir, err = ioutil.TempDir("", "in-destination")
		Ω(err).ShouldNot(HaveOccurred())

		destination = path.Join(tmpdir, "in-dir")

		checkCmd = exec.Command(checkPath, destination)
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		var request models.CheckRequest
		var response models.CheckResponse

		var bucket *s3.Bucket

		BeforeEach(func() {
			guid, err := uuid.NewV4()
			Ω(err).ShouldNot(HaveOccurred())

			key = guid.String()

			auth := aws.Auth{
				AccessKey: accessKeyID,
				SecretKey: secretAccessKey,
			}

			region, ok := aws.Regions[regionName]
			Ω(ok).Should(BeTrue())

			client := s3.New(auth, region)

			bucket = client.Bucket(bucketName)

			request = models.CheckRequest{
				Version: models.Version{},
				Source: models.Source{
					Bucket:          bucketName,
					Key:             key,
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
					RegionName:      regionName,
				},
			}

			response = models.CheckResponse{}
		})

		AfterEach(func() {
			err := bucket.Del(key)
			Ω(err).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			stdin, err := checkCmd.StdinPipe()
			Ω(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())

			// account for roundtrip to s3
			Eventually(session, 5*time.Second).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("with no version", func() {
			BeforeEach(func() {
				request.Version.Number = ""
			})

			Context("when a version is present in the source", func() {
				BeforeEach(func() {
					err := bucket.Put(key, []byte("1.2.3"), "text/plain", "")
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("returns the version present at the source", func() {
					Ω(response).Should(HaveLen(1))
					Ω(response[0].Number).Should(Equal("1.2.3"))
				})
			})

			Context("when no version is present at the source", func() {
				Context("and an initial version is set", func() {
					BeforeEach(func() {
						request.Source.InitialVersion = "10.9.8"
					})

					It("returns the initial version", func() {
						Ω(response).Should(HaveLen(1))
						Ω(response[0].Number).Should(Equal("10.9.8"))
					})
				})

				Context("and an initial version is not set", func() {
					BeforeEach(func() {
						request.Source.InitialVersion = ""
					})

					It("returns the initial version as 0.0.0", func() {
						Ω(response).Should(HaveLen(1))
						Ω(response[0].Number).Should(Equal("0.0.0"))
					})
				})
			})
		})

		Context("with a version present", func() {
			BeforeEach(func() {
				request.Version.Number = "1.2.3"
			})

			Context("when there is no current version", func() {
				It("outputs an empty list", func() {
					Ω(response).Should(HaveLen(0))
				})
			})

			Context("when the source has a higher version", func() {
				BeforeEach(func() {
					err := bucket.Put(key, []byte("1.2.4"), "text/plain", "")
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("returns the version present at the source", func() {
					Ω(response).Should(HaveLen(1))
					Ω(response[0].Number).Should(Equal("1.2.4"))
				})
			})

			Context("when it's the same as the current version", func() {
				BeforeEach(func() {
					err := bucket.Put(key, []byte("1.2.3"), "text/plain", "")
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("outputs an empty list", func() {
					Ω(response).Should(HaveLen(0))
				})
			})
		})
	})
})
