package version

import (
	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
)

func Bump(v semver.Version, params models.InParams) {
	if len(params.Pre) > 0 {
		if len(v.Pre) == 2 {
			if !v.Pre[0].IsNum && v.Pre[0].VersionStr == params.Pre && v.Pre[1].IsNumeric() {
				v.Pre[1].VersionNum++
			} else {
				v.Pre[0] = semver.PRVersion{
					VersionStr: params.Pre,
				}

				v.Pre[1] = semver.PRVersion{
					VersionNum: 1,
					IsNum:      true,
				}
			}
		} else {
			bump(v, params.Bump)

			v.Pre = []semver.PRVersion{
				{VersionStr: params.Pre},
				{VersionNum: 1, IsNum: true},
			}
		}
	} else {
		bump(v, params.Bump)
	}
}

func bump(v semver.Version, t string) {
	switch t {
	case "major":
		v.Major++
		v.Minor = 0
		v.Patch = 0
		v.Pre = nil
	case "minor":
		v.Minor++
		v.Patch = 0
		v.Pre = nil
	case "patch":
		v.Patch++
		v.Pre = nil
	case "final":
		v.Pre = nil
	}
}
