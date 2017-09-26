package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/concourse/semver-resource/models"
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
		Expect(err).NotTo(HaveOccurred())

		destination = path.Join(tmpdir, "in-dir")

		checkCmd = exec.Command(checkPath, destination)
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		var request models.CheckRequest
		var response models.CheckResponse
		var svc *s3.S3

		BeforeEach(func() {
			guid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())

			key = guid.String()

			creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
			awsConfig := &aws.Config{
				Region:           aws.String(regionName),
				Credentials:      creds,
				S3ForcePathStyle: aws.Bool(true),
				MaxRetries:       aws.Int(12),
			}

			svc = s3.New(session.New(awsConfig))

			request = models.CheckRequest{
				Version: models.Version{},
				Source: models.Source{
					Bucket:          bucketName,
					Key:             key,
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
					RegionName:      regionName,
					UseV2Signing:    v2signing,
				},
			}

			response = models.CheckResponse{}
		})

		AfterEach(func() {
			_, err := svc.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			stdin, err := checkCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(checkCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			// account for roundtrip to s3
			Eventually(session, 5*time.Second).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		})

		putVersion := func(version string) {
			_, err := svc.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(bucketName),
				Key:         aws.String(key),
				ContentType: aws.String("text/plain"),
				Body:        bytes.NewReader([]byte(version)),
				ACL:         aws.String(s3.ObjectCannedACLPrivate),
			})
			Expect(err).NotTo(HaveOccurred())
		}

		Context("with no version", func() {
			BeforeEach(func() {
				request.Version.Number = ""
			})

			Context("when a version is present in the source", func() {
				BeforeEach(func() {
					putVersion("1.2.3")
				})

				It("returns the version present at the source", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Number).To(Equal("1.2.3"))
				})
			})

			Context("when no version is present at the source", func() {
				Context("and an initial version is set", func() {
					BeforeEach(func() {
						request.Source.InitialVersion = "10.9.8"
					})

					It("returns the initial version", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Number).To(Equal("10.9.8"))
					})
				})

				Context("and an initial version is not set", func() {
					BeforeEach(func() {
						request.Source.InitialVersion = ""
					})

					It("returns the initial version as 0.0.0", func() {
						Expect(response).To(HaveLen(1))
						Expect(response[0].Number).To(Equal("0.0.0"))
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
					Expect(response).To(HaveLen(0))
				})
			})

			Context("when the source has a higher version", func() {
				BeforeEach(func() {
					putVersion("1.2.4")
				})

				It("returns the version present at the source", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Number).To(Equal("1.2.4"))
				})
			})

			Context("when it's the same as the current version", func() {
				BeforeEach(func() {
					putVersion("1.2.3")
				})

				It("returns the version present at the source", func() {
					Expect(response).To(HaveLen(1))
					Expect(response[0].Number).To(Equal("1.2.3"))
				})
			})
		})
	})
})
