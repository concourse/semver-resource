package version

import "github.com/blang/semver"

type BuildBump struct {
	Build []string
}

func (bump BuildBump) Apply(v semver.Version) semver.Version {
	v.Build = append(v.Build, bump.Build...)

	return v
}

