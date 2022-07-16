package telegraph

import (
	"bytes"
	"context"
	"image"
	"image/draw"
	"image/png"
	"testing"

	"arhat.dev/mbot/pkg/rt"
	"github.com/stretchr/testify/assert"
)

func TestDriver(t *testing.T) {
	var (
		d Driver

		fakeData bytes.Buffer
	)

	img := image.NewAlpha(image.Rect(0, 0, 50, 50))
	draw.Draw(img, image.Rect(0, 0, 0, 0), image.White, image.Point{}, draw.Src)
	err := png.Encode(&fakeData, img)
	assert.NoError(t, err)

	input := rt.NewInput(int64(fakeData.Len()), &fakeData)
	url, err := d.Upload(context.TODO(), "", rt.NewMIME("image/png"), &input)
	assert.NoError(t, err)
	t.Log(url)
}
