package driver_test

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/driver"
	. "github.com/onsi/ginkgo/v2"
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
			Expect(s.params.ServerSideEncryption).To(Equal(types.ServerSideEncryption("my-encryption-schema")))
		})
		It("leaves it empty when disabled", func() {
			s := &service{}
			d := driver.S3Driver{
				Svc: s,
			}
			d.Set(semver.Version{})
			Expect(s.params.ServerSideEncryption).To(BeEmpty())
		})
	})
})

type service struct {
	params *s3.PutObjectInput
}

func (*service) GetObject(ctx context.Context, p *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return nil, nil
}

func (s *service) PutObject(ctx context.Context, p *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	s.params = p
	return nil, nil
}
