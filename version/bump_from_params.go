package version

func BumpFromParams(bumpStr string, preStr string, preWithoutVersion bool, buildStr string, BuildWithoutVersion bool) Bump {
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
		bump = append(bump, PreBump{preStr, preWithoutVersion})
	}

	if buildStr != "" {
		bump = append(bump, BuildBump{buildStr, BuildWithoutVersion})
	}
	return bump
}
