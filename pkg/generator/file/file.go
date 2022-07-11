package file

import (
	"bytes"
	"path"

	"arhat.dev/pkg/fshelper"
	"arhat.dev/rs"

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

func (c *Config) Create() (_ generator.Interface, err error) {
	fs := fshelper.NewOSFS(false, func() (string, error) {
		return c.Dir, nil
	})

	err = fs.MkdirAll(".", 0755)
	if err != nil {
		return
	}

	return &Driver{fs: fs}, nil
}

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	fs *fshelper.OSFS
}

func (d *Driver) Name() string { return Name }

// RenderPageHeader does nothing for this driver
func (d *Driver) RenderPageHeader() ([]byte, error) {
	return nil, nil
}

// RenderPageBody saves all multi-media to local file, return filenames of them in bytes, separated by '\n'
// one filename each line
//
// non multi-media entities (links and plain text) are not touched
func (d *Driver) RenderPageBody(messages []message.Interface) (_ []byte, err error) {
	var (
		buf bytes.Buffer
	)

	for _, msg := range messages {
		for _, e := range msg.Entities() {
			if e.SpanFlags&message.SpanFlagsColl_MultiMedia == 0 {
				continue
			}

			if e.Data == nil {
				continue
			}

			var (
				fileExt     string
				oldFilename string
			)

			if !e.Filename.IsNil() {
				oldFilename = e.Filename.Get()
				fileExt = path.Ext(oldFilename)
			}

			if len(fileExt) == 0 {
				if !e.URL.IsNil() {
					fileExt = path.Ext(e.URL.Get())
				}
			}

			filename := e.Data.Name() + fileExt
			// TODO: write
			// err2 := os.WriteFile(filepath.Join(d.dir, filename), data, 0644)
			// if err2 != nil {
			// 	err = multierr.Append(err, err2)
			// }

			buf.WriteString(filename)
			buf.WriteByte('\n')
		}
	}

	return buf.Next(buf.Len()), err
}
