package file

import (
	"bytes"
	"encoding/hex"
	"os"
	"path"
	"path/filepath"

	"arhat.dev/pkg/sha256helper"
	"arhat.dev/rs"
	"go.uber.org/multierr"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

// nolint:revive
const (
	Name = "file"
)

func init() {
	generator.Register(
		Name,
		func() generator.Config { return &Config{} },
	)
}

type Config struct {
	rs.BaseField

	Dir string `json:"dir" yaml:"dir"`
}

func (c *Config) Create() (generator.Interface, error) {
	return &Driver{dir: c.Dir}, nil
}

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	dir string
}

func (d *Driver) Name() string { return Name }

func (d *Driver) RenderPageHeader() ([]byte, error) {
	return nil, nil
}

// RenderPageBody saves all multi-media to local file, return filenames of them in bytes, separated by '\n'
// one filename each line
//
// non multi-media entities (links and plain text) are not touched
func (d *Driver) RenderPageBody(messages []message.Interface) (_ []byte, err error) {
	err = os.MkdirAll(d.dir, 0750)
	if err != nil && !os.IsExist(err) {
		return
	}

	var buf bytes.Buffer
	for _, msg := range messages {
		for _, e := range msg.Entities() {
			if e.Kind&(message.KindAudio|
				message.KindVideo|
				message.KindImage|
				message.KindFile) == 0 {
				continue
			}

			if e.Params == nil {
				continue
			}

			dataVal, ok := e.Params[message.EntityParamData]
			if !ok {
				continue
			}

			data, ok := dataVal.([]byte)
			if !ok {
				continue
			}

			fileExt := ""
			oldFilenameVal, ok := e.Params[message.EntityParamFilename]
			if ok {
				oldFilename, ok := oldFilenameVal.(string)
				if ok {
					fileExt = path.Ext(oldFilename)
				}
			}

			if fileExt == "" {
				urlVal, ok := e.Params[message.EntityParamURL]
				if ok {
					url, ok := urlVal.(string)
					if ok {
						fileExt = path.Ext(url)
					}
				}
			}

			filename := hex.EncodeToString(sha256helper.Sum(data)) + fileExt
			err2 := os.WriteFile(filepath.Join(d.dir, filename), data, 0644)
			if err2 != nil {
				err = multierr.Append(err, err2)
			}

			buf.WriteString(filename)
			buf.WriteByte('\n')
		}
	}

	return buf.Next(buf.Len()), err
}
