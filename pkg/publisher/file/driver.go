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
func (d *Driver) AppendToExisting(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) (_ []rt.Span, err error) {
	filename := normalizeFilename(d.currentFilename.Load().(string))
	f, err := os.OpenFile(filepath.Join(d.dir, filename), os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return nil, err
	}

	defer func() { _ = f.Close() }()

	_, err = f.ReadFrom(fromGenerator.Reader())
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

// CreateNew implements publisher.Interface
func (d *Driver) CreateNew(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) (_ []rt.Span, err error) {
	filename := normalizeFilename(params)
	d.currentFilename.Store(filename)

	f, err := os.OpenFile(filename, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return
	}

	defer f.Close()

	_, err = f.ReadFrom(fromGenerator.Reader())
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

func (d *Driver) RequireLogin(con rt.Conversation, cmd, params string) (rt.LoginFlow, error) {
	return rt.LoginFlow_None, nil
}

func (d *Driver) Login(con rt.Conversation, user publisher.User) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) RequestExternalAccess(con rt.Conversation) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List(con rt.Conversation) ([]publisher.PostInfo, error) {
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

func (d *Driver) Delete(con rt.Conversation, cmd, params string) error {
	var err error
	keys := strings.Split(params, " ")
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

func normalizeFilename(title string) string {
	return title
}
