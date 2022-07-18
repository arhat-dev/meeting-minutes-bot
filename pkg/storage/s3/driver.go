package s3

import (
	"fmt"
	"path"

	"github.com/minio/minio-go/v7"

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

func (s *Driver) Upload(con rt.Conversation, in *rt.StorageInput) (out rt.StorageOutput, err error) {
	if len(s.bucket) != 0 {
		hasBucket, err2 := s.client.BucketExists(con.Context(), s.bucket)
		if err2 != nil {
			err = fmt.Errorf("check bucket existence: %w", err2)
			return
		}

		if !hasBucket {
			err = s.client.MakeBucket(con.Context(), s.bucket, minio.MakeBucketOptions{
				Region: s.region,
			})
			if err != nil {
				err = fmt.Errorf("create bucket: %w", err)
				return
			}
		}
	}

	objectKey := path.Join(s.basePath, in.Filename())
	info, err := s.client.PutObject(
		con.Context(),
		s.bucket,
		objectKey,
		in.Reader(),
		in.Size(),
		minio.PutObjectOptions{
			ContentType: in.ContentType(),
		},
	)
	if err != nil {
		err = fmt.Errorf("put object: %w", err)
		return
	}

	// we cannot use presign url, since max expiry time is 7 days
	// so if you want this file accessible from browser, update your
	// bucket settings
	out.URL = s.formatPublicURL(con.Context(), info.Key)
	return
}
