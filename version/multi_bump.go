package version

import "github.com/blang/semver"

type MultiBump []Bump

func (bumps MultiBump) Apply(v semver.Version) semver.Version {
	for _, bump := range bumps {
		v = bump.Apply(v)
	}

	return v
}
