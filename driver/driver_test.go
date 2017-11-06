package driver_test

import (
	"net/http"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/concourse/semver-resource/driver"
	"github.com/concourse/semver-resource/models"
	. "github.com/onsi/ginkgo"
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
			svc, ok := s3Driver.Svc.(*s3.S3)
			Expect(ok).To(BeTrue())
			Expect(svc.Client.Config.HTTPClient).Should(BeEquivalentTo(http.DefaultClient))
		})
		It("returns a s3 driver with a transport that ignores ssl verification", func() {
			src.SkipSSLVerification = true
			aDriver, err := driver.FromSource(src)
			Expect(err).To(BeNil())
			Expect(aDriver).ToNot(BeNil())
			s3Driver, ok := aDriver.(*driver.S3Driver)
			Expect(ok).To(BeTrue())
			Expect(s3Driver.Svc).To(Not(BeNil()))
			svc, ok := s3Driver.Svc.(*s3.S3)
			Expect(ok).To(BeTrue())
			Expect(svc.Client.Config.HTTPClient.Transport).ToNot(BeNil())
			transport, ok := svc.Client.Config.HTTPClient.Transport.(*http.Transport)
			Expect(ok).To(BeTrue())
			Expect(transport.TLSClientConfig.InsecureSkipVerify).Should(BeTrue())
		})
	})
})
