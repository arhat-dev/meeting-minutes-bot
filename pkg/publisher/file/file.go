package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"

	"go.uber.org/multierr"

	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

// nolint:revive
const (
	Name = "file"
)

func init() {
	publisher.Register(
		Name,
		func(config interface{}) (publisher.Interface, publisher.UserConfig, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, nil, fmt.Errorf("unexpected non file config")
			}

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
		},
		func() interface{} {
			return &Config{}
		},
	)
}

type Config struct {
	Dir string `json:"dir" yaml:"dir"`
}

var _ publisher.UserConfig = (*UserConfig)(nil)

type UserConfig struct{}

func (u *UserConfig) SetAuthToken(token string) {}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	dir string

	currentFilename *atomic.Value
}

func (d *Driver) Name() string {
	return Name
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

func (d *Driver) Retrieve(key string) ([]message.Entity, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List() ([]publisher.PostInfo, error) {
	entries, err := ioutil.ReadDir(d.dir)
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

func (d *Driver) Publish(title string, body []byte) ([]message.Entity, error) {
	filename := normalizeFilename(title)
	d.currentFilename.Store(filename)

	return []message.Entity{
		{
			Kind: message.KindText,
			Text: "Your messages will be rendered into ",
		},
		{
			Kind: message.KindPre,
			Text: filename,
		},
	}, os.WriteFile(filepath.Join(d.dir, filename), body, 0640)
}

func (d *Driver) Append(body []byte) ([]message.Entity, error) {
	filename := normalizeFilename(d.currentFilename.Load().(string))
	f, err := os.OpenFile(filepath.Join(d.dir, filename), os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	_, err = f.Write(body)
	if err != nil {
		return nil, err
	}

	return []message.Entity{
		{
			Kind: message.KindText,
			Text: "Your messages have been rendered into ",
		},
		{
			Kind: message.KindPre,
			Text: filename,
		},
	}, nil
}

func normalizeFilename(title string) string {
	return title
}
