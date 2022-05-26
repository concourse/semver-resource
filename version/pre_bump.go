package version

import "github.com/blang/semver"

type PreBump struct {
	Pre               string
	PreWithoutVersion bool
}

func (bump PreBump) Apply(v semver.Version) semver.Version {
	if bump.PreWithoutVersion {
		v.Pre = []semver.PRVersion{
			{VersionStr: bump.Pre},
		}
	} else if v.Pre == nil || len(v.Pre) < 2 || v.Pre[0].VersionStr != bump.Pre {
		// no pre-release, no pre-release version, or different pre-release id -> set version to 1
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
