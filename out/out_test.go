package main_test

import (
	"encoding/json"
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
		Ω(err).ShouldNot(HaveOccurred())

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
			Ω(err).ShouldNot(HaveOccurred())

			key = guid.String()

			auth := aws.Auth{
				AccessKey: accessKeyID,
				SecretKey: secretAccessKey,
			}

			client := s3.New(auth, aws.USEast)

			bucket = client.Bucket(bucketName)

			request = models.OutRequest{
				Version: models.Version{},
				Source: models.Source{
					Bucket:          bucketName,
					Key:             key,
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
				},
				Params: models.OutParams{
					File: "number",
				},
			}

			response = models.OutResponse{}
		})

		AfterEach(func() {
			err := bucket.Del(key)
			Ω(err).ShouldNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			stdin, err := outCmd.StdinPipe()
			Ω(err).ShouldNot(HaveOccurred())

			session, err := gexec.Start(outCmd, GinkgoWriter, GinkgoWriter)
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Ω(err).ShouldNot(HaveOccurred())

			// account for roundtrip to s3
			Eventually(session, 5*time.Second).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when a valid version is in the file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(source, "number"), []byte("1.2.3"), 0644)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("reports the version as the resource's version", func() {
				Ω(response.Version.Number).Should(Equal("1.2.3"))
			})

			It("saves the contents of the file in the configured bucket", func() {
				contents, err := bucket.Get(key)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(string(contents)).Should(Equal("1.2.3"))
			})
		})
	})
})
