package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
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

var _ = Describe("In", func() {
	var key string

	var tmpdir string
	var destination string

	var inCmd *exec.Cmd

	BeforeEach(func() {
		var err error

		tmpdir, err = os.MkdirTemp("", "in-destination")
		Expect(err).NotTo(HaveOccurred())

		destination = path.Join(tmpdir, "in-dir")

		inCmd = exec.Command(inPath, destination)
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		var request models.InRequest
		var response models.InResponse

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

			request = models.InRequest{
				Version: models.Version{
					Number: "1.2.3",
				},
				Source: models.Source{
					Bucket:          bucketName,
					Key:             key,
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
					RegionName:      regionName,
				},
				Params: models.InParams{},
			}

			response = models.InResponse{}
		})

		AfterEach(func() {
			_, err := svc.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			stdin, err := inCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(inCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			// account for roundtrip to s3
			Eventually(session, 5*time.Second).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
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

				It("reports the original version as the version", func() {
					Expect(response.Version.Number).To(Equal(request.Version.Number))
				})

				It("writes the version to the destination 'number' file", func() {
					contents, err := os.ReadFile(path.Join(destination, "number"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal(resultLocal))
				})

				It("writes the version to the destination 'version' file", func() {
					contents, err := os.ReadFile(path.Join(destination, "version"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal(resultLocal))
				})
			})
		}
	})
})
