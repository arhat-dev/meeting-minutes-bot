package s3

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/minio/minio-go/v7"

	"arhat.dev/meeting-minutes-bot/pkg/storage"
)

// nolint:revive
const (
	Name = "s3"
)

func init() {
	storage.Register(
		Name,
		func() storage.Config { return &Config{} },
	)
}

var _ storage.Interface = (*Driver)(nil)

type Driver struct {
	client *minio.Client

	bucket   string
	region   string
	basePath string
}

func (s *Driver) Name() string { return Name }

func (s *Driver) Upload(
	ctx context.Context, filename, contentType string, size int64, data io.Reader,
) (url string, err error) {
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
		data,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to put object: %w", err)
	}

	// we cannot use presign url, since max expiry time is 7 days
	// so if you want this file accessible from browser, update your
	// bucket settings
	return s.formatPublicURL(ctx, info.Key), nil
}
