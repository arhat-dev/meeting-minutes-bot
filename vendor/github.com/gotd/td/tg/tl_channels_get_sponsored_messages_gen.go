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

// ChannelsGetSponsoredMessagesRequest represents TL type `channels.getSponsoredMessages#ec210fbf`.
// Get a list of sponsored messages
//
// See https://core.telegram.org/method/channels.getSponsoredMessages for reference.
type ChannelsGetSponsoredMessagesRequest struct {
	// Peer
	Channel InputChannelClass
}

// ChannelsGetSponsoredMessagesRequestTypeID is TL type id of ChannelsGetSponsoredMessagesRequest.
const ChannelsGetSponsoredMessagesRequestTypeID = 0xec210fbf

// Ensuring interfaces in compile-time for ChannelsGetSponsoredMessagesRequest.
var (
	_ bin.Encoder     = &ChannelsGetSponsoredMessagesRequest{}
	_ bin.Decoder     = &ChannelsGetSponsoredMessagesRequest{}
	_ bin.BareEncoder = &ChannelsGetSponsoredMessagesRequest{}
	_ bin.BareDecoder = &ChannelsGetSponsoredMessagesRequest{}
)

func (g *ChannelsGetSponsoredMessagesRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Channel == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *ChannelsGetSponsoredMessagesRequest) String() string {
	if g == nil {
		return "ChannelsGetSponsoredMessagesRequest(nil)"
	}
	type Alias ChannelsGetSponsoredMessagesRequest
	return fmt.Sprintf("ChannelsGetSponsoredMessagesRequest%+v", Alias(*g))
}

// FillFrom fills ChannelsGetSponsoredMessagesRequest from given interface.
func (g *ChannelsGetSponsoredMessagesRequest) FillFrom(from interface {
	GetChannel() (value InputChannelClass)
}) {
	g.Channel = from.GetChannel()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*ChannelsGetSponsoredMessagesRequest) TypeID() uint32 {
	return ChannelsGetSponsoredMessagesRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*ChannelsGetSponsoredMessagesRequest) TypeName() string {
	return "channels.getSponsoredMessages"
}

// TypeInfo returns info about TL type.
func (g *ChannelsGetSponsoredMessagesRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "channels.getSponsoredMessages",
		ID:   ChannelsGetSponsoredMessagesRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Channel",
			SchemaName: "channel",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *ChannelsGetSponsoredMessagesRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode channels.getSponsoredMessages#ec210fbf as nil")
	}
	b.PutID(ChannelsGetSponsoredMessagesRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *ChannelsGetSponsoredMessagesRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode channels.getSponsoredMessages#ec210fbf as nil")
	}
	if g.Channel == nil {
		return fmt.Errorf("unable to encode channels.getSponsoredMessages#ec210fbf: field channel is nil")
	}
	if err := g.Channel.Encode(b); err != nil {
		return fmt.Errorf("unable to encode channels.getSponsoredMessages#ec210fbf: field channel: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *ChannelsGetSponsoredMessagesRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode channels.getSponsoredMessages#ec210fbf to nil")
	}
	if err := b.ConsumeID(ChannelsGetSponsoredMessagesRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode channels.getSponsoredMessages#ec210fbf: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *ChannelsGetSponsoredMessagesRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode channels.getSponsoredMessages#ec210fbf to nil")
	}
	{
		value, err := DecodeInputChannel(b)
		if err != nil {
			return fmt.Errorf("unable to decode channels.getSponsoredMessages#ec210fbf: field channel: %w", err)
		}
		g.Channel = value
	}
	return nil
}

// GetChannel returns value of Channel field.
func (g *ChannelsGetSponsoredMessagesRequest) GetChannel() (value InputChannelClass) {
	if g == nil {
		return
	}
	return g.Channel
}

// GetChannelAsNotEmpty returns mapped value of Channel field.
func (g *ChannelsGetSponsoredMessagesRequest) GetChannelAsNotEmpty() (NotEmptyInputChannel, bool) {
	return g.Channel.AsNotEmpty()
}

// ChannelsGetSponsoredMessages invokes method channels.getSponsoredMessages#ec210fbf returning error if any.
// Get a list of sponsored messages
//
// See https://core.telegram.org/method/channels.getSponsoredMessages for reference.
func (c *Client) ChannelsGetSponsoredMessages(ctx context.Context, channel InputChannelClass) (*MessagesSponsoredMessages, error) {
	var result MessagesSponsoredMessages

	request := &ChannelsGetSponsoredMessagesRequest{
		Channel: channel,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
