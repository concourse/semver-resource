package version

import "github.com/blang/semver"

type PreBump struct {
	Pre string
	PreWithoutVersion bool
}

func (bump PreBump) Apply(v semver.Version) semver.Version {
	if bump.PreWithoutVersion {
		v.Pre = []semver.PRVersion{
			{VersionStr: bump.Pre},
		}
	} else if v.Pre == nil || v.Pre[0].VersionStr != bump.Pre {
		v.Pre = []semver.PRVersion{
			{VersionStr: bump.Pre},
			{VersionNum: 1, IsNum: true},
		}
	} else {
		if len(v.Pre) < 2 {
			v.Pre = append(v.Pre, semver.PRVersion{VersionNum: 0, IsNum: true})
		}

		v.Pre = []semver.PRVersion{
			{VersionStr: bump.Pre},
			{VersionNum: v.Pre[1].VersionNum + 1, IsNum: true},
		}
	}

	return v
}
