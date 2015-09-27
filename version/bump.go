package version

import "github.com/blang/semver"

type Bump interface {
	Apply(semver.Version) semver.Version
}

type IdentityBump struct{}

func (IdentityBump) Apply(v semver.Version) semver.Version {
	return v
}
