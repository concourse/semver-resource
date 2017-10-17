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

	Bucket               string `json:"bucket"`
	Key                  string `json:"key"`
	AccessKeyID          string `json:"access_key_id"`
	SecretAccessKey      string `json:"secret_access_key"`
	RegionName           string `json:"region_name"`
	Endpoint             string `json:"endpoint"`
	DisableSSL           bool   `json:"disable_ssl"`
	SkipSSLVerification  bool   `json:"skip_ssl_verification"`
	ServerSideEncryption string `json:"server_side_encryption"`
	UseV2Signing         bool   `json:"use_v2_signing"`

	URI        string `json:"uri"`
	Branch     string `json:"branch"`
	PrivateKey string `json:"private_key"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	File       string `json:"file"`
	GitUser    string `json:"git_user"`

	OpenStack OpenStackOptions `json:"openstack"`

	JSONKey string `json:"json_key"`
}

// OpenStackOptions contains properties for authenticating and accessing
// the object storage system.
type OpenStackOptions struct {
	Container string `json:"container"`
	ItemName  string `json:"item_name"`
	Region    string `json:"region"`

	// Properties below are for authentication. Its a copy of
	// the properties required by gophercloud. Review documentation
	// in gophercloud for parameter usage as these are just passed in.
	IdentityEndpoint string `json:"identity_endpoint"`
	Username         string `json:"username"`
	UserID           string `json:"user_id"`
	Password         string `json:"password"`
	APIKey           string `json:"api_key"`
	DomainID         string `json:"domain_id"`
	DomainName       string `json:"domain_name"`
	TenantID         string `json:"tenant_id"`
	TenantName       string `json:"tenant_name"`
	AllowReauth      bool   `json:"allow_reauth"`
	TokenID          string `json:"token_id"`
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
	DriverSwift       Driver = "swift"
	DriverGCS         Driver = "gcs"
)
