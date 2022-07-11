// Code generated by gotdgen, DO NOT EDIT.

package tg

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.uber.org/multierr"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdjson"
	"github.com/gotd/td/tdp"
	"github.com/gotd/td/tgerr"
)

// No-op definition for keeping imports.
var (
	_ = bin.Buffer{}
	_ = context.Background()
	_ = fmt.Stringer(nil)
	_ = strings.Builder{}
	_ = errors.Is
	_ = multierr.AppendInto
	_ = sort.Ints
	_ = tdp.Format
	_ = tgerr.Error{}
	_ = tdjson.Encoder{}
)

// WebViewResultURL represents TL type `webViewResultUrl#c14557c`.
//
// See https://core.telegram.org/constructor/webViewResultUrl for reference.
type WebViewResultURL struct {
	// QueryID field of WebViewResultURL.
	QueryID int64
	// URL field of WebViewResultURL.
	URL string
}

// WebViewResultURLTypeID is TL type id of WebViewResultURL.
const WebViewResultURLTypeID = 0xc14557c

// Ensuring interfaces in compile-time for WebViewResultURL.
var (
	_ bin.Encoder     = &WebViewResultURL{}
	_ bin.Decoder     = &WebViewResultURL{}
	_ bin.BareEncoder = &WebViewResultURL{}
	_ bin.BareDecoder = &WebViewResultURL{}
)

func (w *WebViewResultURL) Zero() bool {
	if w == nil {
		return true
	}
	if !(w.QueryID == 0) {
		return false
	}
	if !(w.URL == "") {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (w *WebViewResultURL) String() string {
	if w == nil {
		return "WebViewResultURL(nil)"
	}
	type Alias WebViewResultURL
	return fmt.Sprintf("WebViewResultURL%+v", Alias(*w))
}

// FillFrom fills WebViewResultURL from given interface.
func (w *WebViewResultURL) FillFrom(from interface {
	GetQueryID() (value int64)
	GetURL() (value string)
}) {
	w.QueryID = from.GetQueryID()
	w.URL = from.GetURL()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*WebViewResultURL) TypeID() uint32 {
	return WebViewResultURLTypeID
}

// TypeName returns name of type in TL schema.
func (*WebViewResultURL) TypeName() string {
	return "webViewResultUrl"
}

// TypeInfo returns info about TL type.
func (w *WebViewResultURL) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "webViewResultUrl",
		ID:   WebViewResultURLTypeID,
	}
	if w == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "QueryID",
			SchemaName: "query_id",
		},
		{
			Name:       "URL",
			SchemaName: "url",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (w *WebViewResultURL) Encode(b *bin.Buffer) error {
	if w == nil {
		return fmt.Errorf("can't encode webViewResultUrl#c14557c as nil")
	}
	b.PutID(WebViewResultURLTypeID)
	return w.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (w *WebViewResultURL) EncodeBare(b *bin.Buffer) error {
	if w == nil {
		return fmt.Errorf("can't encode webViewResultUrl#c14557c as nil")
	}
	b.PutLong(w.QueryID)
	b.PutString(w.URL)
	return nil
}

// Decode implements bin.Decoder.
func (w *WebViewResultURL) Decode(b *bin.Buffer) error {
	if w == nil {
		return fmt.Errorf("can't decode webViewResultUrl#c14557c to nil")
	}
	if err := b.ConsumeID(WebViewResultURLTypeID); err != nil {
		return fmt.Errorf("unable to decode webViewResultUrl#c14557c: %w", err)
	}
	return w.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (w *WebViewResultURL) DecodeBare(b *bin.Buffer) error {
	if w == nil {
		return fmt.Errorf("can't decode webViewResultUrl#c14557c to nil")
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode webViewResultUrl#c14557c: field query_id: %w", err)
		}
		w.QueryID = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode webViewResultUrl#c14557c: field url: %w", err)
		}
		w.URL = value
	}
	return nil
}

// GetQueryID returns value of QueryID field.
func (w *WebViewResultURL) GetQueryID() (value int64) {
	if w == nil {
		return
	}
	return w.QueryID
}

// GetURL returns value of URL field.
func (w *WebViewResultURL) GetURL() (value string) {
	if w == nil {
		return
	}
	return w.URL
}
