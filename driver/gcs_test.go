package driver_test

import (
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/blang/semver"
	. "github.com/concourse/semver-resource/driver"
	"github.com/concourse/semver-resource/version"

	. "github.com/onsi/ginkgo/v2"
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

				Expect(versions).To(HaveLen(1))
				Expect(versions[0]).To(Equal(semver.Version{
					Major: 2,
					Minor: 6,
					Patch: 3,
				}))
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

var _ = Describe("GCS Driver Lifecycle", func() {
	It("handles wrapped ErrObjectNotExist from newer GCS library versions", func() {
		wrappedErr := fmt.Errorf("%w: %w", storage.ErrObjectNotExist, fmt.Errorf("googleapi: Error 404: No such object"))
		wrappedServicer := &FakeIOServicer{
			Buf:      gbytes.NewBuffer(),
			GetError: wrappedErr,
		}

		d := &GCSDriver{
			InitialVersion: semver.Version{Major: 2, Minor: 0, Patch: 0},
			Servicer:       wrappedServicer,
			BucketName:     "test-bucket",
			Key:            "test-key",
		}

		By("Check with nil cursor should return initial_version even with wrapped error")
		versions, err := d.Check(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(versions).To(HaveLen(1))
		Expect(versions[0].String()).To(Equal("2.0.0"))

		By("Check with non-nil cursor should return empty list even with wrapped error")
		cursor := semver.Version{Major: 2, Minor: 0, Patch: 0}
		versions, err = d.Check(&cursor)
		Expect(err).NotTo(HaveOccurred())
		Expect(versions).To(BeEmpty())
	})

	It("simulates Concourse check/put cycle with initial_version and no pre-existing object", func() {
		statefulServicer := &StatefulFakeIOServicer{
			objectExists: false,
			Buf:          gbytes.NewBuffer(),
		}

		d := &GCSDriver{
			InitialVersion: semver.Version{Major: 1, Minor: 0, Patch: 0},
			Servicer:       statefulServicer,
			BucketName:     "test-bucket",
			Key:            "test-key",
		}

		By("First check: no cursor, object doesn't exist -> returns initial_version")
		versions, err := d.Check(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(versions).To(HaveLen(1))
		Expect(versions[0].String()).To(Equal("1.0.0"))

		By("Put step: bump patch -> should write 1.0.1")
		newV, err := d.Bump(version.PatchBump{})
		Expect(err).NotTo(HaveOccurred())
		Expect(newV.String()).To(Equal("1.0.1"))

		By("Second check: cursor=1.0.0, object now exists with 1.0.1 -> returns 1.0.1")
		cursor := semver.Version{Major: 1, Minor: 0, Patch: 0}
		versions, err = d.Check(&cursor)
		Expect(err).NotTo(HaveOccurred())
		Expect(versions).To(HaveLen(1))
		Expect(versions[0].String()).To(Equal("1.0.1"))

		By("Another put step: bump patch -> should write 1.0.2")
		statefulServicer.Buf = gbytes.NewBuffer()
		newV, err = d.Bump(version.PatchBump{})
		Expect(err).NotTo(HaveOccurred())
		Expect(newV.String()).To(Equal("1.0.2"))

		By("Third check: cursor=1.0.1 -> returns 1.0.2")
		cursor = semver.Version{Major: 1, Minor: 0, Patch: 1}
		versions, err = d.Check(&cursor)
		Expect(err).NotTo(HaveOccurred())
		Expect(versions).To(HaveLen(1))
		Expect(versions[0].String()).To(Equal("1.0.2"))
	})

	It("demonstrates the bug: Bump always reads initial_version because Check(nil) ignores stored version when object doesn't exist", func() {
		statefulServicer := &StatefulFakeIOServicer{
			objectExists: false,
			Buf:          gbytes.NewBuffer(),
		}

		d := &GCSDriver{
			InitialVersion: semver.Version{Major: 5, Minor: 0, Patch: 0},
			Servicer:       statefulServicer,
			BucketName:     "test-bucket",
			Key:            "test-key",
		}

		By("First bump: Check(nil) -> initial 5.0.0 -> bump to 5.0.1 -> Set writes 5.0.1")
		newV, err := d.Bump(version.PatchBump{})
		Expect(err).NotTo(HaveOccurred())
		Expect(newV.String()).To(Equal("5.0.1"))
		Expect(statefulServicer.storedVersion).To(Equal("5.0.1"))

		By("Second bump: Check(nil) -> should read 5.0.1 from bucket -> bump to 5.0.2")
		statefulServicer.Buf = gbytes.NewBuffer()
		newV, err = d.Bump(version.PatchBump{})
		Expect(err).NotTo(HaveOccurred())
		Expect(newV.String()).To(Equal("5.0.2"))
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

	return io.NopCloser(strings.NewReader(s.Body)), s.GetError
}

func (s *FakeIOServicer) PutObject(bucketName, objectName string) (io.WriteCloser, error) {
	s.BucketName = bucketName
	s.ObjectName = objectName

	return s.Buf, nil
}

type StatefulFakeIOServicer struct {
	storedVersion string
	objectExists  bool
	Buf           *gbytes.Buffer
}

func (s *StatefulFakeIOServicer) GetObject(bucketName, objectName string) (io.ReadCloser, error) {
	if !s.objectExists {
		return io.NopCloser(strings.NewReader("")), storage.ErrObjectNotExist
	}
	return io.NopCloser(strings.NewReader(s.storedVersion)), nil
}

func (s *StatefulFakeIOServicer) PutObject(bucketName, objectName string) (io.WriteCloser, error) {
	return &statefulWriter{servicer: s, buf: s.Buf}, nil
}

type statefulWriter struct {
	servicer *StatefulFakeIOServicer
	buf      *gbytes.Buffer
	data     []byte
}

func (w *statefulWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return w.buf.Write(p)
}

func (w *statefulWriter) Close() error {
	w.servicer.storedVersion = string(w.data)
	w.servicer.objectExists = true
	return w.buf.Close()
}
