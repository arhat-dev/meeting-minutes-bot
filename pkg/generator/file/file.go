package file

import (
	"bytes"
	"path"

	"arhat.dev/pkg/fshelper"
	"arhat.dev/rs"

	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

// nolint:revive
const (
	Name = "file"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
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
func (d *Driver) RenderPageBody(msgs []*rt.Message) (_ []byte, err error) {
	var (
		buf bytes.Buffer
	)

	for _, m := range msgs {
		for _, e := range m.Spans {
			if e.Flags&rt.SpanFlagColl_Media == 0 {
				continue
			}

			if e.Data == nil {
				continue
			}

			var (
				fileExt     string
				oldFilename string
			)

			if len(e.Filename) != 0 {
				oldFilename = e.Filename
				fileExt = path.Ext(oldFilename)
			}

			if len(fileExt) == 0 {
				if len(e.URL) != 0 {
					fileExt = path.Ext(e.URL)
				}
			}

			filename := e.Data.ID().String() + fileExt
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
