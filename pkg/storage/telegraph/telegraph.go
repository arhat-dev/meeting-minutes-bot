package telegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
)

type Driver struct{ client http.Client }

func (d *Driver) Name() string { return Name }

var quoteUnescaper = strings.NewReplacer("\\", "", "\\\"", `"`)

func unescapeQuotes(s string) string {
	return quoteUnescaper.Replace(s)
}

func (d *Driver) Upload(
	ctx context.Context, filename, contentType string, size int64, data io.Reader,
) (url string, err error) {
	var (
		body bytes.Buffer
	)

	_ = filename
	mw := multipart.NewWriter(&body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="blob"`)
	h.Set("Content-Type", contentType)
	filePart, err := mw.CreatePart(h)
	if err != nil {
		return "", fmt.Errorf("failed to prepare file write: %w", err)
	}

	_, err = io.Copy(filePart, data)
	if err != nil {
		return "", fmt.Errorf("failed to buffer request body: %w", err)
	}

	err = mw.Close()
	if err != nil {
		return "", fmt.Errorf("multipart form closed with error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://telegra.ph/upload", &body)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://telegra.ph/")
	req.Header.Set("Origin", "https://telegra.ph")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request file upload: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("file upload failed: %v", string(respBody))
	}

	type uploadResp struct {
		SRC string `json:"src"`
	}

	var result []uploadResp
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", fmt.Errorf("parse upload response %q: %w", string(respBody), err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("no url returned")
	}

	url = "https://telegra.ph/" + strings.TrimLeft(unescapeQuotes(result[0].SRC), "/")
	return url, nil
}
