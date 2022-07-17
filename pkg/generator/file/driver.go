package file

import (
	"path"
	"strings"

	"arhat.dev/pkg/fshelper"

	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	fs *fshelper.OSFS
}

// New implements generator.Interface
func (*Driver) New(con rt.Conversation, cmd, params string) (string, error) {
	return "", nil
}

// Continue implements generator.Interface
func (*Driver) Continue(con rt.Conversation, cmd string, params string) (string, error) {
	return "", nil
}

// RenderBody implements generator.Interface
//
// it saves all multi-media to local file, return filenames of them in bytes, separated by '\n'
// one filename each line
//
// non multi-media entities (links and plain text) are not touched
func (*Driver) RenderBody(con rt.Conversation, msgs []*rt.Message) (_ string, err error) {
	var (
		buf strings.Builder
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

	return buf.String(), err
}
