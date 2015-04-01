package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"

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

	auth := aws.Auth{
		AccessKey: request.Source.AccessKeyID,
		SecretKey: request.Source.SecretAccessKey,
	}

	regionName := request.Source.RegionName
	if len(regionName) == 0 {
		regionName = aws.USEast.Name
	}

	region, ok := aws.Regions[regionName]
	if !ok {
		fatal("resolving region name", errors.New(fmt.Sprintf("No such region '%s'", regionName)))
	}

	client := s3.New(auth, region)
	bucket := client.Bucket(request.Source.Bucket)

	versionNumber := request.Version.Number
	if len(versionNumber) == 0 {
		bucketNumber, err := bucket.Get(request.Source.Key)
		if err == nil {
			versionNumber = string(bucketNumber)
		} else if s3err, ok := err.(*s3.Error); ok && s3err.StatusCode == 404 {
			versionNumber = "0.0.0"
		} else {
			fatal("fetching current version", err)
		}
	}

	inVersion := request.Version
	inVersion.Number = versionNumber

	v, err := semver.Parse(versionNumber)
	if err != nil {
		fatal("parsing semantic version", err)
	}

	version.Bump(v, request.Params)

	numberFile, err := os.Create(filepath.Join(destination, "number"))
	if err != nil {
		fatal("opening number file", err)
	}

	defer numberFile.Close()

	_, err = fmt.Fprintf(numberFile, "%s", v.String())
	if err != nil {
		fatal("writing to number file", err)
	}

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
		Version: inVersion,
		Metadata: models.Metadata{
			{"number", inVersion.Number},
		},
	})
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
