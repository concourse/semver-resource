package version

import "github.com/blang/semver"

type MajorBump struct{}

func (MajorBump) Apply(v semver.Version) semver.Version {
	v.Major++
	v.Minor = 0
	v.Patch = 0
	v.Pre = nil
	return v
}
