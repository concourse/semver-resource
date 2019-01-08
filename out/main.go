package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"

	"github.com/concourse/semver-resource/driver"
	"github.com/concourse/semver-resource/models"
	"github.com/concourse/semver-resource/version"
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

	if request.Params.DriverReadOnly != nil {
		request.Source.DriverReadOnly = *request.Params.DriverReadOnly
	}

	driver, err := driver.FromSource(request.Source)
	if err != nil {
		fatal("constructing driver", err)
	}

	var newVersion semver.Version
	if request.Params.File != "" {
		versionFile, err := os.Open(filepath.Join(sources, request.Params.File))
		if err != nil {
			fatal("opening version file", err)
		}

		defer versionFile.Close()

		var versionStr string
		_, err = fmt.Fscanf(versionFile, "%s", &versionStr)
		if err != nil {
			fatal("reading version file", err)
		}

		newVersion, err = semver.Parse(versionStr)
		if err != nil {
			fatal("parsing version", err)
		}

		err = driver.Set(newVersion)
		if err != nil {
			fatal("setting version", err)
		}
	} else if request.Params.BumpFile != "" || request.Params.PreFile != "" {
		var bumpStr string
		var preStr string

		if request.Params.BumpFile != "" {
			pathToFile := request.Params.BumpFile
			if strings.Index(pathToFile, "/") != 0 {
				pathToFile = filepath.Join(sources, pathToFile)
			}
			bumpStr = readFromFile(pathToFile, "bump")
		}

		if request.Params.PreFile != "" {
			pathToFile := request.Params.PreFile
			if strings.Index(pathToFile, "/") != 0 {
				pathToFile = filepath.Join(sources, pathToFile)
			}
			preStr = readFromFile(pathToFile, "pre")
		}
		bump := version.BumpFromParams(bumpStr, preStr)

		newVersion, err = driver.Bump(bump)
		if err != nil {
			fatal("bumping version", err)
		}
	} else if request.Params.Bump != "" || request.Params.Pre != "" {
		bump := version.BumpFromParams(request.Params.Bump, request.Params.Pre)

		newVersion, err = driver.Bump(bump)
		if err != nil {
			fatal("bumping version", err)
		}
	} else {
		println("no version bump specified")
		os.Exit(1)
	}

	outVersion := models.Version{
		Number: newVersion.String(),
	}

	json.NewEncoder(os.Stdout).Encode(models.OutResponse{
		Version: outVersion,
		Metadata: models.Metadata{
			{"number", outVersion.Number},
		},
	})
}

func readFromFile(fn string, kind string) string {
	file, err := os.Open(fn)
	if err != nil {
		fatal("opening "+kind+" file", err)
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		fatal("stating "+kind+" file", err)
	}
	if info.Size() == 0 {
		return ""
	}

	var str string
	_, err = fmt.Fscanf(file, "%s", &str)
	if err != nil {
		fatal("reading "+kind+" file", err)
	}
	return str
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
