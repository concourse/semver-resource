package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blang/semver"

	"github.com/concourse/semver-resource/models"
	"github.com/concourse/semver-resource/version"
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

	inputVersion, err := semver.Parse(request.Version.Number)
	if err != nil {
		fatal("parsing semantic version", err)
	}

	bumped := version.BumpFromParams(request.Params.Bump, request.Params.Pre).Apply(inputVersion)

	if !bumped.Equals(inputVersion) {
		fmt.Fprintf(os.Stderr, "bumped locally from %s to %s\n", inputVersion, bumped)
	}

	versionFileNames := []string{"number", "version"}

	for _, fileName := range versionFileNames {
		numberFile, err := os.Create(filepath.Join(destination, fileName))
		if err != nil {
			fatal("opening number file", err)
		}

		defer numberFile.Close()

		_, err = fmt.Fprintf(numberFile, "%s", bumped.String())
		if err != nil {
			fatal("writing to number file", err)
		}
	}

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
		Version: request.Version,
		Metadata: models.Metadata{
			{"number", request.Version.Number},
		},
	})
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
