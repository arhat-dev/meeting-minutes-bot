package telegraph

import (
	"encoding/json"
	"fmt"
	"io"

	// "mime/multipart"
	"net/http"
	"strings"

	"arhat.dev/mbot/internal/mime"
	"arhat.dev/mbot/internal/multipart"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/pkg/stringhelper"
)

type Driver struct{ client http.Client }

func (d *Driver) Name() string { return Name }

var quoteUnescaper = strings.NewReplacer("\\", "", "\\\"", `"`)

func unescapeQuotes(s string) string {
	return quoteUnescaper.Replace(s)
}

func (d *Driver) Upload(
	con rt.Conversation,
	filename string,
	contentType mime.MIME,
	in *rt.Input,
) (url string, err error) {
	var (
		hb multipart.HeaderBuilder
		pb multipart.Builder
	)

	_ = filename

	switch contentType.Type() {
	case mime.MIMEType_Video:
	case mime.MIMEType_Audio:
	case mime.MIMEType_Image:
	default:
		// TODO: fake it as a png file?
		err = fmt.Errorf("unsupported content type %q", contentType.Value)
		return
	}

	multipartContentType, body := pb.CreatePart(
		hb.Add("Content-Disposition", `form-data; name="file"; filename="blob"`).
			Add("Content-Type", contentType.Value).Build(),
		in.Reader(),
	).Build()

	req, err := http.NewRequestWithContext(con.Context(), http.MethodPost, "https://telegra.ph/upload", &body)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", multipartContentType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://telegra.ph/")
	req.Header.Set("Origin", "https://telegra.ph")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request file upload: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read response: %w", err)
		}

		return "", fmt.Errorf(
			"failed to upload file, code %d: %s",
			resp.StatusCode,
			stringhelper.Convert[string, byte](respBody),
		)
	}

	var result UploadResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&result)
	if err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}

	if len(result.Err.Error) != 0 {
		err = errString(result.Err.Error)
		return
	}

	if len(result.Sources) == 0 {
		err = errString("unexpected no return url")
		return
	}

	url = "https://telegra.ph/" + strings.TrimLeft(unescapeQuotes(result.Sources[0].Src), "/")
	return
}

type UploadResponse struct {
	Sources []struct {
		Src string `json:"src"`
	}
	Err struct {
		Error string `json:"error"`
	}
}

func (r *UploadResponse) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	switch data[0] {
	case '{':
		return json.Unmarshal(data, &r.Err)
	case '[':
		return json.Unmarshal(data, &r.Sources)
	default:
		return fmt.Errorf("unexpected response data: %s", stringhelper.Convert[string, byte](data))
	}
}

type errString string

func (e errString) Error() string { return string(e) }
