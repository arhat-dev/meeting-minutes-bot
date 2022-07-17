package telegraph

import (
	"bytes"
	"context"
	"image"
	"image/draw"
	"image/png"
	"testing"

	"arhat.dev/mbot/internal/mime"
	"arhat.dev/mbot/pkg/rt"
	rttest "arhat.dev/mbot/pkg/rt/test"
	"github.com/stretchr/testify/assert"
)

func TestDriver(t *testing.T) {
	var (
		d Driver

		fakeData bytes.Buffer
	)

	con := rttest.FakeConversation(context.TODO())

	img := image.NewAlpha(image.Rect(0, 0, 50, 50))
	draw.Draw(img, image.Rect(0, 0, 0, 0), image.White, image.Point{}, draw.Src)
	err := png.Encode(&fakeData, img)
	assert.NoError(t, err)

	input := rt.NewInput(int64(fakeData.Len()), &fakeData)
	url, err := d.Upload(con, "", mime.New("image/png"), &input)
	assert.NoError(t, err)
	t.Log(url)
}
