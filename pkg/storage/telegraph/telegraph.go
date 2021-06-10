package telegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"arhat.dev/meeting-minutes-bot/pkg/storage"
)

const Name = "telegraph"

func init() {
	storage.Register(
		Name,
		func(config interface{}) (storage.Interface, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, fmt.Errorf("unexpected non telegraph storage config")
			}
			_ = c

			d := &Driver{
				client: &http.Client{},
			}
			return d, nil
		},
		func() interface{} {
			return &Config{}
		},
	)
}

type Config struct {
}

type Driver struct {
	client *http.Client
}

func (d *Driver) Name() string {
	return Name
}

var quoteUnescaper = strings.NewReplacer("\\", "", "\\\"", `"`)

func unescapeQuotes(s string) string {
	return quoteUnescaper.Replace(s)
}

func (d *Driver) Upload(
	ctx context.Context,
	filename string,
	contentType string,
	data []byte,
) (url string, err error) {
	_ = filename

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="blob"`)
	h.Set("Content-Type", contentType)
	filePart, err := mw.CreatePart(h)
	if err != nil {
		return "", fmt.Errorf("failed to prepare file write: %w", err)
	}

	_, err = filePart.Write(data)
	if err != nil {
		return "", fmt.Errorf("failed to buffer request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://telegra.ph/upload", body)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request file upload: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("file upload failed: %v", string(respData))
	}

	type uploadResp struct {
		SRC string `json:"src"`
	}

	result := &uploadResp{}
	err = json.Unmarshal(respData, result)
	if err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}

	return unescapeQuotes(result.SRC), nil
}
