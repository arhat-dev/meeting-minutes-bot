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

// StatsGetMessageStatsRequest represents TL type `stats.getMessageStats#b6e0a3f5`.
// Get message statistics¹
//
// Links:
//  1) https://core.telegram.org/api/stats
//
// See https://core.telegram.org/method/stats.getMessageStats for reference.
type StatsGetMessageStatsRequest struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Whether to enable dark theme for graph colors
	Dark bool
	// Channel ID
	Channel InputChannelClass
	// Message ID
	MsgID int
}

// StatsGetMessageStatsRequestTypeID is TL type id of StatsGetMessageStatsRequest.
const StatsGetMessageStatsRequestTypeID = 0xb6e0a3f5

// Ensuring interfaces in compile-time for StatsGetMessageStatsRequest.
var (
	_ bin.Encoder     = &StatsGetMessageStatsRequest{}
	_ bin.Decoder     = &StatsGetMessageStatsRequest{}
	_ bin.BareEncoder = &StatsGetMessageStatsRequest{}
	_ bin.BareDecoder = &StatsGetMessageStatsRequest{}
)

func (g *StatsGetMessageStatsRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Flags.Zero()) {
		return false
	}
	if !(g.Dark == false) {
		return false
	}
	if !(g.Channel == nil) {
		return false
	}
	if !(g.MsgID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *StatsGetMessageStatsRequest) String() string {
	if g == nil {
		return "StatsGetMessageStatsRequest(nil)"
	}
	type Alias StatsGetMessageStatsRequest
	return fmt.Sprintf("StatsGetMessageStatsRequest%+v", Alias(*g))
}

// FillFrom fills StatsGetMessageStatsRequest from given interface.
func (g *StatsGetMessageStatsRequest) FillFrom(from interface {
	GetDark() (value bool)
	GetChannel() (value InputChannelClass)
	GetMsgID() (value int)
}) {
	g.Dark = from.GetDark()
	g.Channel = from.GetChannel()
	g.MsgID = from.GetMsgID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*StatsGetMessageStatsRequest) TypeID() uint32 {
	return StatsGetMessageStatsRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*StatsGetMessageStatsRequest) TypeName() string {
	return "stats.getMessageStats"
}

// TypeInfo returns info about TL type.
func (g *StatsGetMessageStatsRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "stats.getMessageStats",
		ID:   StatsGetMessageStatsRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Dark",
			SchemaName: "dark",
			Null:       !g.Flags.Has(0),
		},
		{
			Name:       "Channel",
			SchemaName: "channel",
		},
		{
			Name:       "MsgID",
			SchemaName: "msg_id",
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (g *StatsGetMessageStatsRequest) SetFlags() {
	if !(g.Dark == false) {
		g.Flags.Set(0)
	}
}

// Encode implements bin.Encoder.
func (g *StatsGetMessageStatsRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode stats.getMessageStats#b6e0a3f5 as nil")
	}
	b.PutID(StatsGetMessageStatsRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *StatsGetMessageStatsRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode stats.getMessageStats#b6e0a3f5 as nil")
	}
	g.SetFlags()
	if err := g.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode stats.getMessageStats#b6e0a3f5: field flags: %w", err)
	}
	if g.Channel == nil {
		return fmt.Errorf("unable to encode stats.getMessageStats#b6e0a3f5: field channel is nil")
	}
	if err := g.Channel.Encode(b); err != nil {
		return fmt.Errorf("unable to encode stats.getMessageStats#b6e0a3f5: field channel: %w", err)
	}
	b.PutInt(g.MsgID)
	return nil
}

// Decode implements bin.Decoder.
func (g *StatsGetMessageStatsRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode stats.getMessageStats#b6e0a3f5 to nil")
	}
	if err := b.ConsumeID(StatsGetMessageStatsRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode stats.getMessageStats#b6e0a3f5: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *StatsGetMessageStatsRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode stats.getMessageStats#b6e0a3f5 to nil")
	}
	{
		if err := g.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode stats.getMessageStats#b6e0a3f5: field flags: %w", err)
		}
	}
	g.Dark = g.Flags.Has(0)
	{
		value, err := DecodeInputChannel(b)
		if err != nil {
			return fmt.Errorf("unable to decode stats.getMessageStats#b6e0a3f5: field channel: %w", err)
		}
		g.Channel = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode stats.getMessageStats#b6e0a3f5: field msg_id: %w", err)
		}
		g.MsgID = value
	}
	return nil
}

// SetDark sets value of Dark conditional field.
func (g *StatsGetMessageStatsRequest) SetDark(value bool) {
	if value {
		g.Flags.Set(0)
		g.Dark = true
	} else {
		g.Flags.Unset(0)
		g.Dark = false
	}
}

// GetDark returns value of Dark conditional field.
func (g *StatsGetMessageStatsRequest) GetDark() (value bool) {
	if g == nil {
		return
	}
	return g.Flags.Has(0)
}

// GetChannel returns value of Channel field.
func (g *StatsGetMessageStatsRequest) GetChannel() (value InputChannelClass) {
	if g == nil {
		return
	}
	return g.Channel
}

// GetMsgID returns value of MsgID field.
func (g *StatsGetMessageStatsRequest) GetMsgID() (value int) {
	if g == nil {
		return
	}
	return g.MsgID
}

// GetChannelAsNotEmpty returns mapped value of Channel field.
func (g *StatsGetMessageStatsRequest) GetChannelAsNotEmpty() (NotEmptyInputChannel, bool) {
	return g.Channel.AsNotEmpty()
}

// StatsGetMessageStats invokes method stats.getMessageStats#b6e0a3f5 returning error if any.
// Get message statistics¹
//
// Links:
//  1) https://core.telegram.org/api/stats
//
// Possible errors:
//  400 CHANNEL_INVALID: The provided channel is invalid.
//  400 CHAT_ADMIN_REQUIRED: You must be an admin in this chat to do this.
//
// See https://core.telegram.org/method/stats.getMessageStats for reference.
func (c *Client) StatsGetMessageStats(ctx context.Context, request *StatsGetMessageStatsRequest) (*StatsMessageStats, error) {
	var result StatsMessageStats

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}