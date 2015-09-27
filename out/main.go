package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blang/semver"

	"github.com/concourse/semver-resource/driver"
	"github.com/concourse/semver-resource/models"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <source>")
		os.Exit(1)
	}

	sources := os.Args[1]

	var request models.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("reading request", err)
	}

	contents, err := ioutil.ReadFile(filepath.Join(sources, request.Params.File))
	if err != nil {
		fatal("reading version file", err)
	}

	v, err := semver.Parse(string(contents))
	if err != nil {
		fatal("parsing version", err)
	}

	driver, err := driver.FromSource(request.Source)
	if err != nil {
		fatal("constructing driver", err)
	}

	err = driver.Set(v)
	if err != nil {
		fatal("setting version", err)
	}

	outVersion := models.Version{
		Number: v.String(),
	}

	json.NewEncoder(os.Stdout).Encode(models.OutResponse{
		Version: outVersion,
		Metadata: models.Metadata{
			{"number", outVersion.Number},
		},
	})
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
