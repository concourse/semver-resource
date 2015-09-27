package version

func BumpFromParams(bumpStr string, preStr string) Bump {
	var semverBump Bump

	switch bumpStr {
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
	if semverBump != nil && preStr != "" {
		bump = ConditionalPreBump{
			ConditionalBump: semverBump,
			Pre:             preStr,
		}
	} else if preStr != "" {
		bump = PreBump{preStr}
	} else if semverBump != nil {
		bump = semverBump
	} else {
		bump = IdentityBump{}
	}

	return bump
}
