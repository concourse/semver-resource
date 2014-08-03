package main

import (
	"encoding/json"
	"os"

	"github.com/concourse/semver-resource/models"
)

func main() {
	json.NewEncoder(os.Stdout).Encode([]models.Version{})
}
