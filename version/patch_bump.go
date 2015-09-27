package version

import "github.com/blang/semver"

type PatchBump struct{}

func (PatchBump) Apply(v semver.Version) semver.Version {
	v.Patch++
	v.Pre = nil
	return v
}
