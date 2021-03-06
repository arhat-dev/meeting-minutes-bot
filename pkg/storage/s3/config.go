package s3

import (
	"fmt"
	"net/url"

	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/rs"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// nolint:revive
const (
	Name = "s3"
)

func init() {
	storage.Register(Name, func() storage.Config { return &Config{} })
}

// Config for s3 file uploader
type Config struct {
	rs.BaseField

	EndpointURL string `yaml:"endpointURL"`
	Region      string `yaml:"region"`

	Bucket   string `yaml:"bucket"`
	BasePath string `yaml:"basePath"`

	AccessKeyID     string `yaml:"accessKeyID"`
	AccessKeySecret string `yaml:"accessKeySecret"`
}

func (c *Config) Create() (storage.Interface, error) {
	eURL, err := url.Parse(c.EndpointURL)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint url: %w", err)
	}

	client, err := minio.New(eURL.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKeyID, c.AccessKeySecret, ""),
		Secure: eURL.Scheme == "https",
		Region: c.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %w", err)
	}

	return &Driver{
		client: client,

		bucket:   c.Bucket,
		region:   c.Region,
		basePath: c.BasePath,
	}, nil
}
