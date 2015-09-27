package version

import "github.com/blang/semver"

type FinalBump struct{}

func (FinalBump) Apply(v semver.Version) semver.Version {
	v.Pre = nil
	return v
}
