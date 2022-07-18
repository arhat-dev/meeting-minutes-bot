package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"go.uber.org/multierr"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	dir string

	currentFilename atomic.Value
}

// AppendToExisting implements publisher.Interface
func (d *Driver) AppendToExisting(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	filename := normalizeFilename(d.currentFilename.Load().(string))
	f, err := os.OpenFile(filepath.Join(d.dir, filename), os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return
	}

	defer func() { _ = f.Close() }()

	_, err = f.WriteString(in.Data.Get())
	if err != nil {
		return
	}

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{
				Flags: rt.SpanFlag_PlainText,
				Text:  "Your messages have been rendered into ",
			},
			{
				Flags: rt.SpanFlag_Pre,
				Text:  filename,
			},
		},
	})
	return
}

// CreateNew implements publisher.Interface
func (d *Driver) CreateNew(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	filename := normalizeFilename(params)
	d.currentFilename.Store(filename)

	f, err := os.OpenFile(filename, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return
	}

	defer f.Close()

	_, err = f.WriteString(in.Data.Get())
	if err != nil {
		return
	}

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{
				Flags: rt.SpanFlag_PlainText,
				Text:  "Your messages will be rendered into ",
			},
			{
				Flags: rt.SpanFlag_Code,
				Text:  filename,
			},
		},
	})

	return
}

func (d *Driver) RequireLogin(con rt.Conversation, cmd, params string) (rt.LoginFlow, error) {
	return rt.LoginFlow_None, nil
}

func (d *Driver) Login(con rt.Conversation, user publisher.User) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) RequestExternalAccess(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) List(con rt.Conversation) (out rt.PublisherOutput, err error) {
	entries, err := os.ReadDir(d.dir)
	if err != nil {
		return
	}
	var result rt.SendMessageOptions
	for _, f := range entries {
		if f.IsDir() {
			continue
		}

		result.Body = append(result.Body,
			rt.Span{Text: "- "},
			rt.Span{
				Flags: rt.SpanFlag_Code,
				Text:  f.Name(),
			},
			rt.Span{Text: "\n"},
		)
	}
	out.SendMessage.Set(result)
	return
}

func (d *Driver) Delete(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	keys := strings.Split(params, " ")
	for _, filename := range keys {
		path := filepath.Join(d.dir, filename)
		if filepath.Dir(path) != d.dir {
			err = multierr.Append(err, fmt.Errorf("invalid filename with path"))
			continue
		}

		err = multierr.Append(err, os.Remove(path))
	}

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{Text: "File(s) deleted."},
		},
	})

	return
}

func normalizeFilename(title string) string {
	return title
}
