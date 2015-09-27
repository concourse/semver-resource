package version

import "github.com/blang/semver"

type PreBump struct {
	Pre string
}

func (bump PreBump) Apply(v semver.Version) semver.Version {
	if v.Pre == nil || v.Pre[0].VersionStr != bump.Pre {
		v.Pre = []semver.PRVersion{
			{VersionStr: bump.Pre},
			{VersionNum: 1, IsNum: true},
		}
	} else {
		v.Pre = []semver.PRVersion{
			{VersionStr: bump.Pre},
			{VersionNum: v.Pre[1].VersionNum + 1, IsNum: true},
		}
	}

	return v
}
