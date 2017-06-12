package driver_test

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 Driver", func() {
	Context("with encryption", func() {
		It("sets it when enabled", func() {
			s := &service{}
			d := driver.S3Driver{
				Svc:                  s,
				ServerSideEncryption: "my-encryption-schema",
			}
			d.Set(semver.Version{})
			Expect(*s.params.ServerSideEncryption).To(Equal("my-encryption-schema"))
		})
		It("leaves it empty when disabled", func() {
			s := &service{}
			d := driver.S3Driver{
				Svc: s,
			}
			d.Set(semver.Version{})
			Expect(s.params.ServerSideEncryption).To(BeNil())
		})
	})
})

type service struct {
	params *s3.PutObjectInput
}

func (*service) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return nil, nil
}

func (s *service) PutObject(p *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	s.params = p
	return nil, nil
}
