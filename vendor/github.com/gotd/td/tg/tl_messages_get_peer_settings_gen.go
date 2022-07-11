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

// MessagesGetPeerSettingsRequest represents TL type `messages.getPeerSettings#efd9a6a2`.
// Get peer settings
//
// See https://core.telegram.org/method/messages.getPeerSettings for reference.
type MessagesGetPeerSettingsRequest struct {
	// The peer
	Peer InputPeerClass
}

// MessagesGetPeerSettingsRequestTypeID is TL type id of MessagesGetPeerSettingsRequest.
const MessagesGetPeerSettingsRequestTypeID = 0xefd9a6a2

// Ensuring interfaces in compile-time for MessagesGetPeerSettingsRequest.
var (
	_ bin.Encoder     = &MessagesGetPeerSettingsRequest{}
	_ bin.Decoder     = &MessagesGetPeerSettingsRequest{}
	_ bin.BareEncoder = &MessagesGetPeerSettingsRequest{}
	_ bin.BareDecoder = &MessagesGetPeerSettingsRequest{}
)

func (g *MessagesGetPeerSettingsRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Peer == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *MessagesGetPeerSettingsRequest) String() string {
	if g == nil {
		return "MessagesGetPeerSettingsRequest(nil)"
	}
	type Alias MessagesGetPeerSettingsRequest
	return fmt.Sprintf("MessagesGetPeerSettingsRequest%+v", Alias(*g))
}

// FillFrom fills MessagesGetPeerSettingsRequest from given interface.
func (g *MessagesGetPeerSettingsRequest) FillFrom(from interface {
	GetPeer() (value InputPeerClass)
}) {
	g.Peer = from.GetPeer()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessagesGetPeerSettingsRequest) TypeID() uint32 {
	return MessagesGetPeerSettingsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*MessagesGetPeerSettingsRequest) TypeName() string {
	return "messages.getPeerSettings"
}

// TypeInfo returns info about TL type.
func (g *MessagesGetPeerSettingsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messages.getPeerSettings",
		ID:   MessagesGetPeerSettingsRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Peer",
			SchemaName: "peer",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *MessagesGetPeerSettingsRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode messages.getPeerSettings#efd9a6a2 as nil")
	}
	b.PutID(MessagesGetPeerSettingsRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *MessagesGetPeerSettingsRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode messages.getPeerSettings#efd9a6a2 as nil")
	}
	if g.Peer == nil {
		return fmt.Errorf("unable to encode messages.getPeerSettings#efd9a6a2: field peer is nil")
	}
	if err := g.Peer.Encode(b); err != nil {
		return fmt.Errorf("unable to encode messages.getPeerSettings#efd9a6a2: field peer: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *MessagesGetPeerSettingsRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode messages.getPeerSettings#efd9a6a2 to nil")
	}
	if err := b.ConsumeID(MessagesGetPeerSettingsRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode messages.getPeerSettings#efd9a6a2: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *MessagesGetPeerSettingsRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode messages.getPeerSettings#efd9a6a2 to nil")
	}
	{
		value, err := DecodeInputPeer(b)
		if err != nil {
			return fmt.Errorf("unable to decode messages.getPeerSettings#efd9a6a2: field peer: %w", err)
		}
		g.Peer = value
	}
	return nil
}

// GetPeer returns value of Peer field.
func (g *MessagesGetPeerSettingsRequest) GetPeer() (value InputPeerClass) {
	if g == nil {
		return
	}
	return g.Peer
}

// MessagesGetPeerSettings invokes method messages.getPeerSettings#efd9a6a2 returning error if any.
// Get peer settings
//
// Possible errors:
//  400 CHANNEL_INVALID: The provided channel is invalid.
//  400 PEER_ID_INVALID: The provided peer id is invalid.
//
// See https://core.telegram.org/method/messages.getPeerSettings for reference.
func (c *Client) MessagesGetPeerSettings(ctx context.Context, peer InputPeerClass) (*MessagesPeerSettings, error) {
	var result MessagesPeerSettings

	request := &MessagesGetPeerSettingsRequest{
		Peer: peer,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
