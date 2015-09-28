package models

type Version struct {
	Number string `json:"number"`
}

type InRequest struct {
	Source  Source   `json:"source"`
	Version Version  `json:"version"`
	Params  InParams `json:"params"`
}

type InResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type InParams struct {
	Bump string `json:"bump"`
	Pre  string `json:"pre"`
}

type OutRequest struct {
	Source  Source    `json:"source"`
	Version Version   `json:"version"`
	Params  OutParams `json:"params"`
}

type OutResponse struct {
	Version  Version  `json:"version"`
	Metadata Metadata `json:"metadata"`
}

type OutParams struct {
	File string `json:"file"`

	Bump string `json:"bump"`
	Pre  string `json:"pre"`
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version

type Source struct {
	Driver Driver `json:"driver"`

	InitialVersion string `json:"initial_version"`

	Bucket          string `json:"bucket"`
	Key             string `json:"key"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	RegionName      string `json:"region_name"`
	Endpoint        string `json:"endpoint"`

	URI        string `json:"uri"`
	Branch     string `json:"branch"`
	PrivateKey string `json:"private_key"`
	File       string `json:"file"`
}

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Driver string

const (
	DriverUnspecified Driver = ""
	DriverS3          Driver = "s3"
	DriverGit         Driver = "git"
)
