package s3

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"arhat.dev/meeting-minutes-bot/pkg/storage"
)

// nolint:revive
const (
	Name = "s3"
)

func init() {
	storage.Register(
		Name,
		New,
		func() interface{} {
			return &Config{}
		},
	)
}

// Config for s3 file uploader
type Config struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	Region   string `json:"region" yaml:"region"`

	Bucket   string `json:"bucket" yaml:"bucket"`
	BasePath string `json:"basePath" yaml:"basePath"`

	AccessKeyID     string `json:"accessKeyID" yaml:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret" yaml:"accessKeySecret"`
}

var _ storage.Interface = (*S3)(nil)

func New(config interface{}) (storage.Interface, error) {
	c, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected non s3 config: %T", config)
	}

	eURL, err := url.Parse(c.Endpoint)
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

	return &S3{
		client: client,

		bucket:   c.Bucket,
		region:   c.Region,
		basePath: c.BasePath,
	}, nil
}

type S3 struct {
	client *minio.Client

	bucket   string
	region   string
	basePath string
}

func (s *S3) Name() string {
	return Name
}

func (s *S3) Upload(ctx context.Context, filename string, data []byte) (url string, err error) {
	if len(s.bucket) != 0 {
		hasBucket, err2 := s.client.BucketExists(ctx, s.bucket)
		if err2 != nil {
			return "", fmt.Errorf("failed to check bucket existence: %w", err2)
		}

		if !hasBucket {
			err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
				Region: s.region,
			})
			if err != nil {
				return "", fmt.Errorf("failed to create bucket: %w", err)
			}
		}
	}

	objectKey := path.Join(s.basePath, filename)
	info, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectKey,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("failed to put object: %w", err)
	}

	// we cannot use presign url, since max expiry time is 7 days
	// so if you want this file accessible from browser, update your
	// bucket settings
	return s.formatPublicURL(ctx, info.Key), nil
}
