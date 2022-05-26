package version

import (
	"strconv"

	"github.com/blang/semver"
)

type BuildBump struct {
	Build               string
	BuildWithoutVersion bool
}

func (bump BuildBump) Apply(v semver.Version) semver.Version {
	if bump.BuildWithoutVersion {
		v.Build = []string{
			bump.Build,
		}
	} else if v.Build == nil || len(v.Build) < 2 || v.Build[0] != bump.Build {
		// no build, no build version, different build id -> set build to 1
		v.Build = []string{
			bump.Build, strconv.Itoa(1),
		}
	} else {
		version, err := strconv.Atoi(v.Build[1])
		if err != nil {
			// invalid build version -> set build to 1
			v.Build = []string{
				bump.Build, strconv.Itoa(1),
			}
		}

		v.Build = []string{
			bump.Build, strconv.Itoa(version + 1),
		}
	}

	return v
}
