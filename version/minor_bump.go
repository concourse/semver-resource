package version

import "github.com/blang/semver"

type MinorBump struct{}

func (MinorBump) Apply(v semver.Version) semver.Version {
	v.Minor++
	v.Patch = 0
	v.Pre = nil
	return v
}
