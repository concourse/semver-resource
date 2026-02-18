package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/concourse/semver-resource/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Out", func() {
	var key string

	var source string

	var outCmd *exec.Cmd

	BeforeEach(func() {
		var err error

		source, err = os.MkdirTemp("", "out-source")
		Expect(err).NotTo(HaveOccurred())

		outCmd = exec.Command(outPath, source)
	})

	AfterEach(func() {
		os.RemoveAll(source)
	})

	Context("when executed", func() {
		var request models.OutRequest
		var response models.OutResponse

		var svc *s3.Client

		BeforeEach(func() {
			guid, err := uuid.NewRandom()
			Expect(err).NotTo(HaveOccurred())

			key = guid.String()

			creds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
			cfg, err := config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(regionName),
				config.WithRetryMaxAttempts(12),
				config.WithCredentialsProvider(creds),
			)
			Expect(err).NotTo(HaveOccurred())
			svc = s3.NewFromConfig(cfg, func(o *s3.Options) {
				o.UsePathStyle = true
			})

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
			_, err := svc.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
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

		getVersion := func() string {
			resp, err := svc.GetObject(context.TODO(), &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
			Expect(err).NotTo(HaveOccurred())

			contents, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			return string(contents)
		}

		putVersion := func(version string) {
			_, err := svc.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket:      aws.String(bucketName),
				Key:         aws.String(key),
				ContentType: aws.String("text/plain"),
				Body:        bytes.NewReader([]byte(version)),
			})
			Expect(err).NotTo(HaveOccurred())
		}

		Context("when setting the version", func() {
			BeforeEach(func() {
				request.Params.File = "number"
			})

			Context("when a valid version is in the file", func() {
				BeforeEach(func() {
					err := os.WriteFile(filepath.Join(source, "number"), []byte("1.2.3"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("reports the version as the resource's version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					Expect(getVersion()).To(Equal("1.2.3"))
				})
			})
		})

		Context("when bumping the version", func() {
			BeforeEach(func() {
				putVersion("1.2.3")
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
						Expect(getVersion()).To(Equal(resultLocal))
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
					putVersion("1.2.3")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.1"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					Expect(getVersion()).To(Equal("1.2.3-alpha.1"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0-alpha.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						Expect(getVersion()).To(Equal("1.3.0-alpha.1"))
					})
				})
			})

			Context("when the version is the same prerelease", func() {
				BeforeEach(func() {
					putVersion("1.2.3-alpha.2")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.3"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					Expect(getVersion()).To(Equal("1.2.3-alpha.3"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0-alpha.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						Expect(getVersion()).To(Equal("1.3.0-alpha.1"))
					})
				})
			})

			Context("when the version is a different prerelease", func() {
				BeforeEach(func() {
					putVersion("1.2.3-beta.2")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.1"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					Expect(getVersion()).To(Equal("1.2.3-alpha.1"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0-alpha.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						Expect(getVersion()).To(Equal("1.3.0-alpha.1"))
					})
				})
			})
		})

		Context("when bumping the version to a build", func() {
			BeforeEach(func() {
				request.Params.Build = "build"
			})

			Context("when the version is not a build", func() {
				BeforeEach(func() {
					putVersion("1.2.3")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3+build.1"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					Expect(getVersion()).To(Equal("1.2.3+build.1"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0+build.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						Expect(getVersion()).To(Equal("1.3.0+build.1"))
					})
				})
			})

			Context("when the version is the same build", func() {
				BeforeEach(func() {
					putVersion("1.2.3+build.2")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3+build.3"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					Expect(getVersion()).To(Equal("1.2.3+build.3"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0+build.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						Expect(getVersion()).To(Equal("1.3.0+build.1"))
					})
				})
			})

			Context("when the version is a different build", func() {
				BeforeEach(func() {
					putVersion("1.2.3-beta.2")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3+build.1"))
				})

				It("saves the contents of the file in the configured bucket", func() {
					Expect(getVersion()).To(Equal("1.2.3+build.1"))
				})

				Context("when doing a semantic bump at the same time", func() {
					BeforeEach(func() {
						request.Params.Bump = "minor"
					})

					It("reports the bumped version as the version", func() {
						Expect(response.Version.Number).To(Equal("1.3.0+build.1"))
					})

					It("saves the contents of the file in the configured bucket", func() {
						Expect(getVersion()).To(Equal("1.3.0+build.1"))
					})
				})
			})
		})

		Context("when bumping the version to a prerelease and build", func() {
			BeforeEach(func() {
				request.Params.Pre = "alpha"
				request.Params.Build = "build"
			})

			Context("when the version is just a base version", func() {
				BeforeEach(func() {
					putVersion("1.2.3")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.1+build.1"))
				})
			})

			Context("when the version has a prerelease", func() {
				BeforeEach(func() {
					putVersion("1.2.3-alpha.1")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.2+build.1"))
				})
			})

			Context("when the version has a build", func() {
				BeforeEach(func() {
					putVersion("1.2.3+build.1")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.1+build.2"))
				})
			})

			Context("when the version has a build and prerelease", func() {
				BeforeEach(func() {
					putVersion("1.2.3-alpha.1+build.1")
				})

				It("reports the bumped version as the version", func() {
					Expect(response.Version.Number).To(Equal("1.2.3-alpha.2+build.2"))
				})
			})
		})
	})
})
