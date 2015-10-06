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

	var bump MultiBump
	if semverBump != nil {
		bump = append(bump, semverBump)
	}

	if preStr != "" {
		bump = append(bump, PreBump{preStr})
	}

	return bump
}
