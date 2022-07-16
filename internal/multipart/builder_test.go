package multipart

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	const (
		TEST_DATA     = "test-data"
		TEST_BOUNDARY = "boundary"
	)

	var (
		hdr HeaderBuilder
		b   Builder
	)

	// key order matters
	hdr.Add("Content-Disposition", `form-data; name="file"; filename="blob"`)
	hdr.Add("Content-Type", "application/json")
	assert.NoError(t, b.SetBoundary(TEST_BOUNDARY))
	ct, r := b.
		CreatePart(hdr.Build(), strings.NewReader(TEST_DATA)).
		CreatePart(hdr.Build(), strings.NewReader(TEST_DATA)).
		Build()

	var (
		expected bytes.Buffer
	)
	mw := multipart.NewWriter(&expected)
	assert.NoError(t, mw.SetBoundary(TEST_BOUNDARY))

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="blob"`)
	header.Set("Content-Type", "application/json")

	// first part
	part, err := mw.CreatePart(header)
	assert.NoError(t, err)
	_, err = part.Write([]byte(TEST_DATA))
	assert.NoError(t, err)

	// second part
	part, err = mw.CreatePart(header)
	assert.NoError(t, err)
	_, err = part.Write([]byte(TEST_DATA))
	assert.NoError(t, err)

	mw.Close()

	var (
		actual bytes.Buffer
	)

	_, err = actual.ReadFrom(&r)
	assert.NoError(t, err)
	assert.Equal(t, expected.String(), actual.String())
	assert.Equal(t, mw.FormDataContentType(), ct)
}

func TestRandomBoundary(t *testing.T) {
	b := RandomBoundary()
	t.Log(b)
	assert.Equal(t, 60, len(b))
	assert.NotEqual(t, b, RandomBoundary())
}
