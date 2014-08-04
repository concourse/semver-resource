package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"

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

	auth := aws.Auth{
		AccessKey: request.Source.AccessKeyID,
		SecretKey: request.Source.SecretAccessKey,
	}

	client := s3.New(auth, aws.USEast)
	bucket := client.Bucket(request.Source.Bucket)

	outVersion := models.Version{
		Number: v.String(),
	}

	err = bucket.Put(request.Source.Key, []byte(outVersion.Number), "text/plain", "")
	if err != nil {
		fatal("saving to bucket", err)
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
