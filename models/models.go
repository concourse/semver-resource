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
	Bump  string `json:"bump"`
	Pre   string `json:"pre"`
	Final bool   `json:"final"`
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
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type CheckResponse []Version

type Source struct {
	Bucket          string `json:"bucket"`
	Key             string `json:"key"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	RegionName      string `json:"region_name"`
}

type Metadata []MetadataField

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
