package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/concourse/semver-resource/models"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

func main() {
	var request models.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("reading request", err)
	}

	auth := aws.Auth{
		AccessKey: request.Source.AccessKeyID,
		SecretKey: request.Source.SecretAccessKey,
	}

	regionName := request.Source.RegionName
	region, ok := aws.Regions[regionName]
	if !ok {
		fatal("resolving region name", errors.New(fmt.Sprintf("No such region '%s'", regionName)))
	}

	client := s3.New(auth, region)
	bucket := client.Bucket(request.Source.Bucket)

	var bucketNumber string

	bucketNumberPayload, err := bucket.Get(request.Source.Key)
	if err == nil {
		bucketNumber = string(bucketNumberPayload)
	} else if len(request.Source.InitialVersion) > 0 {
		bucketNumber = request.Source.InitialVersion
	} else {
		bucketNumber = "0.0.0"
	}

	bucketVer, err := semver.Parse(bucketNumber)
	if err != nil {
		fatal("parsing semantic version in bucket", err)
	}

	delta := models.CheckResponse{}
	versionNumber := request.Version.Number
	if len(versionNumber) == 0 {
		delta = append(delta, models.Version{
			Number: bucketNumber,
		})
	} else {
		v, err := semver.Parse(versionNumber)
		if err != nil {
			fatal("parsing semantic version in request", err)
		}

		if bucketVer.GT(v) {
			delta = append(delta, models.Version{
				Number: bucketNumber,
			})
		}
	}

	json.NewEncoder(os.Stdout).Encode(delta)
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
