package version

import "github.com/concourse/semver-resource/models"

func BumpFromParams(params models.InParams) Bump {
	var semverBump Bump

	switch params.Bump {
	case "major":
		semverBump = MajorBump{}
	case "minor":
		semverBump = MinorBump{}
	case "patch":
		semverBump = PatchBump{}
	case "final":
		semverBump = FinalBump{}
	}

	var bump Bump
	if semverBump != nil && params.Pre != "" {
		bump = ConditionalPreBump{
			ConditionalBump: semverBump,
			Pre:             params.Pre,
		}
	} else if params.Pre != "" {
		bump = PreBump{params.Pre}
	} else if semverBump != nil {
		bump = semverBump
	} else {
		bump = IdentityBump{}
	}

	return bump
}
