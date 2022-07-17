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

// PhoneGetCallConfigRequest represents TL type `phone.getCallConfig#55451fa9`.
// Get phone call configuration to be passed to libtgvoip's shared config
//
// See https://core.telegram.org/method/phone.getCallConfig for reference.
type PhoneGetCallConfigRequest struct {
}

// PhoneGetCallConfigRequestTypeID is TL type id of PhoneGetCallConfigRequest.
const PhoneGetCallConfigRequestTypeID = 0x55451fa9

// Ensuring interfaces in compile-time for PhoneGetCallConfigRequest.
var (
	_ bin.Encoder     = &PhoneGetCallConfigRequest{}
	_ bin.Decoder     = &PhoneGetCallConfigRequest{}
	_ bin.BareEncoder = &PhoneGetCallConfigRequest{}
	_ bin.BareDecoder = &PhoneGetCallConfigRequest{}
)

func (g *PhoneGetCallConfigRequest) Zero() bool {
	if g == nil {
		return true
	}

	return true
}

// String implements fmt.Stringer.
func (g *PhoneGetCallConfigRequest) String() string {
	if g == nil {
		return "PhoneGetCallConfigRequest(nil)"
	}
	type Alias PhoneGetCallConfigRequest
	return fmt.Sprintf("PhoneGetCallConfigRequest%+v", Alias(*g))
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PhoneGetCallConfigRequest) TypeID() uint32 {
	return PhoneGetCallConfigRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PhoneGetCallConfigRequest) TypeName() string {
	return "phone.getCallConfig"
}

// TypeInfo returns info about TL type.
func (g *PhoneGetCallConfigRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "phone.getCallConfig",
		ID:   PhoneGetCallConfigRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{}
	return typ
}

// Encode implements bin.Encoder.
func (g *PhoneGetCallConfigRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode phone.getCallConfig#55451fa9 as nil")
	}
	b.PutID(PhoneGetCallConfigRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *PhoneGetCallConfigRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode phone.getCallConfig#55451fa9 as nil")
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *PhoneGetCallConfigRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode phone.getCallConfig#55451fa9 to nil")
	}
	if err := b.ConsumeID(PhoneGetCallConfigRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode phone.getCallConfig#55451fa9: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *PhoneGetCallConfigRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode phone.getCallConfig#55451fa9 to nil")
	}
	return nil
}

// PhoneGetCallConfig invokes method phone.getCallConfig#55451fa9 returning error if any.
// Get phone call configuration to be passed to libtgvoip's shared config
//
// See https://core.telegram.org/method/phone.getCallConfig for reference.
func (c *Client) PhoneGetCallConfig(ctx context.Context) (*DataJSON, error) {
	var result DataJSON

	request := &PhoneGetCallConfigRequest{}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}