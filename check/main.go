package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/driver"
	"github.com/concourse/semver-resource/models"
)

func main() {
	var request models.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("reading request", err)
	}

	driver, err := driver.FromSource(request.Source)
	if err != nil {
		fatal("constructing driver", err)
	}

	var cursor *semver.Version
	if request.Version.Number != "" {
		v, err := semver.Parse(request.Version.Number)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipping invalid current version: %s", err)
		} else {
			cursor = &v
		}
	}

	versions, err := driver.Check(cursor)
	if err != nil {
		fatal("checking for new versions", err)
	}

	delta := models.CheckResponse{}
	for _, v := range versions {
		delta = append(delta, models.Version{
			Number: v.String(),
		})
	}

	json.NewEncoder(os.Stdout).Encode(delta)
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
