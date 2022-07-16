package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"arhat.dev/rs"
	"go.uber.org/multierr"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

// nolint:revive
const (
	Name = "file"
)

func init() {
	publisher.Register(
		Name,
		func() publisher.Config { return &Config{} },
	)
}

type Config struct {
	rs.BaseField

	Dir string `json:"dir" yaml:"dir"`
}

func (c *Config) Create() (publisher.Interface, publisher.UserConfig, error) {
	dir, err := filepath.Abs(c.Dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to determin absolute dir path: %w", err)
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil && !os.IsExist(err) {
		return nil, nil, fmt.Errorf("failed to ensure dir: %w", err)
	}

	return &Driver{
		dir:             dir,
		currentFilename: &atomic.Value{},
	}, &UserConfig{}, nil
}

var _ publisher.UserConfig = (*UserConfig)(nil)

type UserConfig struct{}

func (u *UserConfig) SetAuthToken(token string) {}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	dir string

	currentFilename *atomic.Value
}

func (d *Driver) RequireLogin() bool {
	return false
}

func (d *Driver) Login(config publisher.UserConfig) (token string, _ error) {
	return "", fmt.Errorf("unimplemented")
}

func (d *Driver) AuthURL() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (d *Driver) Retrieve(key string) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List() ([]publisher.PostInfo, error) {
	entries, err := os.ReadDir(d.dir)
	if err != nil {
		return nil, err
	}
	var result []publisher.PostInfo
	for _, f := range entries {
		if f.IsDir() {
			continue
		}

		result = append(result, publisher.PostInfo{
			Title: f.Name(),
			URL:   f.Name(),
		})
	}
	return result, nil
}

func (d *Driver) Delete(keys ...string) error {
	var err error
	for _, filename := range keys {
		path := filepath.Join(d.dir, filename)
		if filepath.Dir(path) != d.dir {
			err = multierr.Append(err, fmt.Errorf("invalid filename with path"))
			continue
		}

		err = multierr.Append(err, os.Remove(path))
	}

	return err
}

func (d *Driver) Publish(title string, body *rt.Input) (_ []rt.Span, err error) {
	filename := normalizeFilename(title)
	d.currentFilename.Store(filename)

	f, err := os.OpenFile(filename, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return
	}

	defer f.Close()

	_, err = f.ReadFrom(body.Reader())
	if err != nil {
		return
	}

	return []rt.Span{
		{
			Flags: rt.SpanFlag_PlainText,
			Text:  "Your messages will be rendered into ",
		},
		{
			Flags: rt.SpanFlag_Code,
			Text:  filename,
		},
	}, nil
}

func (d *Driver) Append(ctx context.Context, body *rt.Input) (_ []rt.Span, err error) {
	filename := normalizeFilename(d.currentFilename.Load().(string))
	f, err := os.OpenFile(filepath.Join(d.dir, filename), os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	_, err = f.ReadFrom(body.Reader())
	if err != nil {
		return nil, err
	}

	return []rt.Span{
		{
			Flags: rt.SpanFlag_PlainText,
			Text:  "Your messages have been rendered into ",
		},
		{
			Flags: rt.SpanFlag_Pre,
			Text:  filename,
		},
	}, nil
}

func normalizeFilename(title string) string {
	return title
}
