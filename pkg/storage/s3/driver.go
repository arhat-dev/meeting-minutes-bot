package s3

import (
	"fmt"
	"path"

	"github.com/minio/minio-go/v7"

	"arhat.dev/mbot/internal/mime"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/storage"
)

var _ storage.Interface = (*Driver)(nil)

type Driver struct {
	client *minio.Client

	bucket   string
	region   string
	basePath string
}

func (s *Driver) Name() string { return Name }

func (s *Driver) Upload(
	con rt.Conversation, filename string, contentType mime.MIME, in *rt.Input,
) (url string, err error) {
	if len(s.bucket) != 0 {
		hasBucket, err2 := s.client.BucketExists(con.Context(), s.bucket)
		if err2 != nil {
			return "", fmt.Errorf("failed to check bucket existence: %w", err2)
		}

		if !hasBucket {
			err = s.client.MakeBucket(con.Context(), s.bucket, minio.MakeBucketOptions{
				Region: s.region,
			})
			if err != nil {
				return "", fmt.Errorf("failed to create bucket: %w", err)
			}
		}
	}

	objectKey := path.Join(s.basePath, filename)
	info, err := s.client.PutObject(
		con.Context(),
		s.bucket,
		objectKey,
		in.Reader(),
		in.Size(),
		minio.PutObjectOptions{
			ContentType: contentType.Value,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to put object: %w", err)
	}

	// we cannot use presign url, since max expiry time is 7 days
	// so if you want this file accessible from browser, update your
	// bucket settings
	return s.formatPublicURL(con.Context(), info.Key), nil
}
