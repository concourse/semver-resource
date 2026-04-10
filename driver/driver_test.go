package driver_test

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/concourse/semver-resource/driver"
	"github.com/concourse/semver-resource/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Driver", func() {
	Context("S3", func() {
		var src models.Source
		BeforeEach(func() {
			src = models.Source{
				Driver: models.DriverS3,
			}
		})
		It("returns a s3 driver with http.DefaultClient", func() {
			aDriver, err := driver.FromSource(src)
			Expect(err).To(BeNil())
			Expect(aDriver).ToNot(BeNil())
			s3Driver, ok := aDriver.(*driver.S3Driver)
			Expect(ok).To(BeTrue())
			Expect(s3Driver.Svc).To(Not(BeNil()))
			svc, ok := s3Driver.Svc.(*s3.Client)
			Expect(ok).To(BeTrue())
			Expect(svc.Options().HTTPClient).Should(BeEquivalentTo(http.DefaultClient))
		})
		It("returns a s3 driver with a transport that ignores ssl verification", func() {
			src.SkipSSLVerification = true
			aDriver, err := driver.FromSource(src)
			Expect(err).To(BeNil())
			Expect(aDriver).ToNot(BeNil())
			s3Driver, ok := aDriver.(*driver.S3Driver)
			Expect(ok).To(BeTrue())
			Expect(s3Driver.Svc).To(Not(BeNil()))
			svc, ok := s3Driver.Svc.(*s3.Client)
			Expect(ok).To(BeTrue())
			httpClient, ok := svc.Options().HTTPClient.(*http.Client)
			Expect(httpClient.Transport).ToNot(BeNil())
			transport, ok := httpClient.Transport.(*http.Transport)
			Expect(ok).To(BeTrue())
			Expect(transport.TLSClientConfig.InsecureSkipVerify).Should(BeTrue())
		})
	})
})

var _ = Describe("Driver", func() {
	Context("Git", func() {
		var src models.Source
		BeforeEach(func() {
			src = models.Source{
				Driver: models.DriverGit,
			}
		})
		It("returns a default git driver", func() {
			aDriver, err := driver.FromSource(src)
			Expect(err).To(BeNil())
			Expect(aDriver).ToNot(BeNil())
			gitDriver, ok := aDriver.(*driver.GitDriver)
			Expect(ok).To(BeTrue())
			Expect(gitDriver.SkipSSLVerification).To(Not(BeNil()))
			Expect(gitDriver.SkipSSLVerification).To(BeFalse())
		})
		It("returns a git driver with a transport that ignores ssl verification", func() {
			src.SkipSSLVerification = true
			aDriver, err := driver.FromSource(src)
			Expect(err).To(BeNil())
			Expect(aDriver).ToNot(BeNil())
			gitDriver, ok := aDriver.(*driver.GitDriver)
			Expect(ok).To(BeTrue())
			Expect(gitDriver.SkipSSLVerification).To(Not(BeNil()))
		})
	})
})

var _ = Describe("Driver", func() {
	Context("GCS", func() {
		var src models.Source
		BeforeEach(func() {
			src = models.Source{
				Driver: models.DriverGCS,
				Bucket: "my-bucket",
				Key:    "my-key",
			}
		})
		It("returns an error when both json_key and token are provided", func() {
			src.JSONKey = `{"type": "service_account"}`
			src.GCSToken = "ya29.some-token"
			_, err := driver.FromSource(src)
			Expect(err).To(MatchError("must specify only one of json_key or token for the gcs driver"))
		})
		It("returns an error when neither json_key nor token is provided", func() {
			_, err := driver.FromSource(src)
			Expect(err).To(MatchError("must specify one of json_key or token for the gcs driver"))
		})
		It("returns a gcs driver when only json_key is provided", func() {
			src.JSONKey = `{"type": "service_account"}`
			aDriver, err := driver.FromSource(src)
			Expect(err).To(BeNil())
			Expect(aDriver).ToNot(BeNil())
			gcsDriver, ok := aDriver.(*driver.GCSDriver)
			Expect(ok).To(BeTrue())
			Expect(gcsDriver.BucketName).To(Equal("my-bucket"))
			Expect(gcsDriver.Key).To(Equal("my-key"))
		})
		It("returns a gcs driver when only token is provided", func() {
			src.GCSToken = "ya29.some-token"
			aDriver, err := driver.FromSource(src)
			Expect(err).To(BeNil())
			Expect(aDriver).ToNot(BeNil())
			gcsDriver, ok := aDriver.(*driver.GCSDriver)
			Expect(ok).To(BeTrue())
			Expect(gcsDriver.BucketName).To(Equal("my-bucket"))
			Expect(gcsDriver.Key).To(Equal("my-key"))
		})
	})
})
