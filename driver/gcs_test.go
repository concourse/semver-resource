package driver_test

import (
	"io"
	"io/ioutil"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/blang/semver"
	. "github.com/concourse/semver-resource/driver"
	"github.com/concourse/semver-resource/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("GCS Driver", func() {
	var (
		buf    *gbytes.Buffer
		s      *FakeIOServicer
		v      semver.Version
		driver *GCSDriver
	)

	BeforeEach(func() {
		buf = gbytes.NewBuffer()

		s = &FakeIOServicer{
			Buf: buf,
		}

		driver = &GCSDriver{
			Servicer:   s,
			BucketName: "fake-bucket",
			Key:        "fake-object",
		}

		v = semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 3,
		}
	})

	Describe("Bump", func() {
		Describe("when the Object exists", func() {
			It("writes the bumped version of the contents back to the object", func() {
				s.Body = "2.6.3"

				newV, err := driver.Bump(version.PatchBump{})

				Expect(err).NotTo(HaveOccurred())
				Expect(newV.String()).To(Equal("2.6.4"))

				Expect(s.BucketName).To(Equal("fake-bucket"))
				Expect(s.ObjectName).To(Equal("fake-object"))
				Expect(s.Buf).To(gbytes.Say("2.6.4"))
			})
		})

		Describe("when the object does not exist", func() {
			It("bumps the initial version", func() {
				s.GetError = storage.ErrObjectNotExist
				driver.InitialVersion = semver.Version{
					Major: 0,
					Minor: 0,
					Patch: 0,
				}

				newV, err := driver.Bump(version.PatchBump{})

				Expect(err).NotTo(HaveOccurred())
				Expect(newV.String()).To(Equal("0.0.1"))
				Expect(s.BucketName).To(Equal("fake-bucket"))
				Expect(s.ObjectName).To(Equal("fake-object"))
				Expect(s.Buf).To(gbytes.Say("0.0.1"))
			})
		})
	})

	Describe("Set", func() {
		It("puts the semver version to the object", func() {
			driver.Set(v)

			Expect(s.BucketName).To(Equal("fake-bucket"))
			Expect(s.ObjectName).To(Equal("fake-object"))

			Expect(buf).To(gbytes.Say(v.String()))
			Expect(buf.Closed()).To(BeTrue())
		})
	})

	Describe("Check", func() {
		Describe("when the object contains a semver version greater than the cursor", func() {
			It("returns the semver version", func() {
				s.Body = "2.6.3"

				versions, err := driver.Check(&semver.Version{
					Major: 1,
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(s.BucketName).To(Equal("fake-bucket"))
				Expect(s.ObjectName).To(Equal("fake-object"))

				Expect(versions).To(HaveLen(1))
				Expect(versions[0]).To(Equal(semver.Version{
					Major: 2,
					Minor: 6,
					Patch: 3,
				}))
			})
		})

		Describe("when the object contains a semver version equal to the cursor", func() {
			It("returns the semver version", func() {
				s.Body = "2.6.3"

				versions, err := driver.Check(&semver.Version{
					Major: 2,
					Minor: 6,
					Patch: 3,
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(s.BucketName).To(Equal("fake-bucket"))
				Expect(s.ObjectName).To(Equal("fake-object"))

				Expect(versions).To(HaveLen(1))
				Expect(versions[0]).To(Equal(semver.Version{
					Major: 2,
					Minor: 6,
					Patch: 3,
				}))
			})
		})

		Describe("when the object contains a semver version less than the cursor", func() {
			It("returns no version", func() {
				s.Body = "2.6.3"

				versions, err := driver.Check(&semver.Version{
					Major: 8,
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(s.BucketName).To(Equal("fake-bucket"))
				Expect(s.ObjectName).To(Equal("fake-object"))

				Expect(versions).To(BeEmpty())
			})
		})

		Describe("when the object contains an invalid string", func() {
			It("returns an error", func() {
				s.Body = "I am not a semver version"

				versions, err := driver.Check(&semver.Version{})

				Expect(versions).To(BeEmpty())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("parsing number in bucket:")))
			})
		})

		Describe("when the object does not exist", func() {
			It("returns the initial version if the cursor does not exist", func() {
				s.GetError = storage.ErrObjectNotExist

				driver.InitialVersion = semver.Version{
					Major: 3,
					Minor: 4,
					Patch: 5,
				}

				versions, err := driver.Check(nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(versions).To(HaveLen(1))
				Expect(versions[0]).To(Equal(driver.InitialVersion))
			})

			It("returns an empty list if the cursor is set", func() {
				s.GetError = storage.ErrObjectNotExist
				driver.InitialVersion = semver.Version{}

				versions, err := driver.Check(&semver.Version{Major: 3})

				Expect(err).NotTo(HaveOccurred())
				Expect(versions).To(BeEmpty())
			})
		})
	})
})

type FakeIOServicer struct {
	Body string
	Buf  *gbytes.Buffer

	BucketName string
	ObjectName string

	GetError error
}

func (s *FakeIOServicer) GetObject(bucketName, objectName string) (io.ReadCloser, error) {
	s.BucketName = bucketName
	s.ObjectName = objectName

	return ioutil.NopCloser(strings.NewReader(s.Body)), s.GetError
}

func (s *FakeIOServicer) PutObject(bucketName, objectName string) (io.WriteCloser, error) {
	s.BucketName = bucketName
	s.ObjectName = objectName

	return s.Buf, nil
}
