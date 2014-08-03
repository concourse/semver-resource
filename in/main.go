package main

import (
	"encoding/json"
	"os"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <destination>")
		os.Exit(1)
	}

	destination := os.Args[1]

	err := os.MkdirAll(destination, 0755)
	if err != nil {
		fatal("creating destination", err)
	}

	var request models.InRequest
	err = json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("reading request", err)
	}

	versionNumber := request.Version.Number
	if len(versionNumber) == 0 {
		versionNumber = "0.0.0"
	}

	v, err := semver.Parse(versionNumber)
	if err != nil {
		fatal("parsing semantic version", err)
	}

	if len(request.Params.Pre) > 0 {
		if len(v.Pre) == 2 {
			if !v.Pre[0].IsNum && v.Pre[0].VersionStr == request.Params.Pre && v.Pre[1].IsNumeric() {
				v.Pre[1].VersionNum++
			} else {
				v.Pre[0] = &semver.PRVersion{
					VersionStr: request.Params.Pre,
				}

				v.Pre[1] = &semver.PRVersion{
					VersionNum: 1,
					IsNum:      true,
				}
			}
		} else {
			bump(v, request.Params.Bump)

			v.Pre = []*semver.PRVersion{
				{VersionStr: request.Params.Pre},
				{VersionNum: 1, IsNum: true},
			}
		}
	} else {
		bump(v, request.Params.Bump)
	}

	inVersion := request.Version
	inVersion.Number = v.String()

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
		Version: inVersion,
		Metadata: models.Metadata{
			{"number", v.String()},
		},
	})
}

func bump(v *semver.Version, t string) {
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

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
