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

// MessageRange represents TL type `messageRange#ae30253`.
// Indicates a range of chat messages
//
// See https://core.telegram.org/constructor/messageRange for reference.
type MessageRange struct {
	// Start of range (message ID)
	MinID int
	// End of range (message ID)
	MaxID int
}

// MessageRangeTypeID is TL type id of MessageRange.
const MessageRangeTypeID = 0xae30253

// Ensuring interfaces in compile-time for MessageRange.
var (
	_ bin.Encoder     = &MessageRange{}
	_ bin.Decoder     = &MessageRange{}
	_ bin.BareEncoder = &MessageRange{}
	_ bin.BareDecoder = &MessageRange{}
)

func (m *MessageRange) Zero() bool {
	if m == nil {
		return true
	}
	if !(m.MinID == 0) {
		return false
	}
	if !(m.MaxID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (m *MessageRange) String() string {
	if m == nil {
		return "MessageRange(nil)"
	}
	type Alias MessageRange
	return fmt.Sprintf("MessageRange%+v", Alias(*m))
}

// FillFrom fills MessageRange from given interface.
func (m *MessageRange) FillFrom(from interface {
	GetMinID() (value int)
	GetMaxID() (value int)
}) {
	m.MinID = from.GetMinID()
	m.MaxID = from.GetMaxID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*MessageRange) TypeID() uint32 {
	return MessageRangeTypeID
}

// TypeName returns name of type in TL schema.
func (*MessageRange) TypeName() string {
	return "messageRange"
}

// TypeInfo returns info about TL type.
func (m *MessageRange) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "messageRange",
		ID:   MessageRangeTypeID,
	}
	if m == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "MinID",
			SchemaName: "min_id",
		},
		{
			Name:       "MaxID",
			SchemaName: "max_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (m *MessageRange) Encode(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't encode messageRange#ae30253 as nil")
	}
	b.PutID(MessageRangeTypeID)
	return m.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (m *MessageRange) EncodeBare(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't encode messageRange#ae30253 as nil")
	}
	b.PutInt(m.MinID)
	b.PutInt(m.MaxID)
	return nil
}

// Decode implements bin.Decoder.
func (m *MessageRange) Decode(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't decode messageRange#ae30253 to nil")
	}
	if err := b.ConsumeID(MessageRangeTypeID); err != nil {
		return fmt.Errorf("unable to decode messageRange#ae30253: %w", err)
	}
	return m.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (m *MessageRange) DecodeBare(b *bin.Buffer) error {
	if m == nil {
		return fmt.Errorf("can't decode messageRange#ae30253 to nil")
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode messageRange#ae30253: field min_id: %w", err)
		}
		m.MinID = value
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode messageRange#ae30253: field max_id: %w", err)
		}
		m.MaxID = value
	}
	return nil
}

// GetMinID returns value of MinID field.
func (m *MessageRange) GetMinID() (value int) {
	if m == nil {
		return
	}
	return m.MinID
}

// GetMaxID returns value of MaxID field.
func (m *MessageRange) GetMaxID() (value int) {
	if m == nil {
		return
	}
	return m.MaxID
}