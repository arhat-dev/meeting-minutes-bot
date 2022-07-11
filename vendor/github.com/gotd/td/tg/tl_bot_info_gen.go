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

// BotInfo represents TL type `botInfo#e4169b5d`.
// Info about bots (available bot commands, etc)
//
// See https://core.telegram.org/constructor/botInfo for reference.
type BotInfo struct {
	// ID of the bot
	UserID int64
	// Description of the bot
	Description string
	// Bot commands that can be used in the chat
	Commands []BotCommand
	// MenuButton field of BotInfo.
	MenuButton BotMenuButtonClass
}

// BotInfoTypeID is TL type id of BotInfo.
const BotInfoTypeID = 0xe4169b5d

// Ensuring interfaces in compile-time for BotInfo.
var (
	_ bin.Encoder     = &BotInfo{}
	_ bin.Decoder     = &BotInfo{}
	_ bin.BareEncoder = &BotInfo{}
	_ bin.BareDecoder = &BotInfo{}
)

func (b *BotInfo) Zero() bool {
	if b == nil {
		return true
	}
	if !(b.UserID == 0) {
		return false
	}
	if !(b.Description == "") {
		return false
	}
	if !(b.Commands == nil) {
		return false
	}
	if !(b.MenuButton == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (b *BotInfo) String() string {
	if b == nil {
		return "BotInfo(nil)"
	}
	type Alias BotInfo
	return fmt.Sprintf("BotInfo%+v", Alias(*b))
}

// FillFrom fills BotInfo from given interface.
func (b *BotInfo) FillFrom(from interface {
	GetUserID() (value int64)
	GetDescription() (value string)
	GetCommands() (value []BotCommand)
	GetMenuButton() (value BotMenuButtonClass)
}) {
	b.UserID = from.GetUserID()
	b.Description = from.GetDescription()
	b.Commands = from.GetCommands()
	b.MenuButton = from.GetMenuButton()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*BotInfo) TypeID() uint32 {
	return BotInfoTypeID
}

// TypeName returns name of type in TL schema.
func (*BotInfo) TypeName() string {
	return "botInfo"
}

// TypeInfo returns info about TL type.
func (b *BotInfo) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "botInfo",
		ID:   BotInfoTypeID,
	}
	if b == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "UserID",
			SchemaName: "user_id",
		},
		{
			Name:       "Description",
			SchemaName: "description",
		},
		{
			Name:       "Commands",
			SchemaName: "commands",
		},
		{
			Name:       "MenuButton",
			SchemaName: "menu_button",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (b *BotInfo) Encode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't encode botInfo#e4169b5d as nil")
	}
	buf.PutID(BotInfoTypeID)
	return b.EncodeBare(buf)
}

// EncodeBare implements bin.BareEncoder.
func (b *BotInfo) EncodeBare(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't encode botInfo#e4169b5d as nil")
	}
	buf.PutLong(b.UserID)
	buf.PutString(b.Description)
	buf.PutVectorHeader(len(b.Commands))
	for idx, v := range b.Commands {
		if err := v.Encode(buf); err != nil {
			return fmt.Errorf("unable to encode botInfo#e4169b5d: field commands element with index %d: %w", idx, err)
		}
	}
	if b.MenuButton == nil {
		return fmt.Errorf("unable to encode botInfo#e4169b5d: field menu_button is nil")
	}
	if err := b.MenuButton.Encode(buf); err != nil {
		return fmt.Errorf("unable to encode botInfo#e4169b5d: field menu_button: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (b *BotInfo) Decode(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't decode botInfo#e4169b5d to nil")
	}
	if err := buf.ConsumeID(BotInfoTypeID); err != nil {
		return fmt.Errorf("unable to decode botInfo#e4169b5d: %w", err)
	}
	return b.DecodeBare(buf)
}

// DecodeBare implements bin.BareDecoder.
func (b *BotInfo) DecodeBare(buf *bin.Buffer) error {
	if b == nil {
		return fmt.Errorf("can't decode botInfo#e4169b5d to nil")
	}
	{
		value, err := buf.Long()
		if err != nil {
			return fmt.Errorf("unable to decode botInfo#e4169b5d: field user_id: %w", err)
		}
		b.UserID = value
	}
	{
		value, err := buf.String()
		if err != nil {
			return fmt.Errorf("unable to decode botInfo#e4169b5d: field description: %w", err)
		}
		b.Description = value
	}
	{
		headerLen, err := buf.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode botInfo#e4169b5d: field commands: %w", err)
		}

		if headerLen > 0 {
			b.Commands = make([]BotCommand, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value BotCommand
			if err := value.Decode(buf); err != nil {
				return fmt.Errorf("unable to decode botInfo#e4169b5d: field commands: %w", err)
			}
			b.Commands = append(b.Commands, value)
		}
	}
	{
		value, err := DecodeBotMenuButton(buf)
		if err != nil {
			return fmt.Errorf("unable to decode botInfo#e4169b5d: field menu_button: %w", err)
		}
		b.MenuButton = value
	}
	return nil
}

// GetUserID returns value of UserID field.
func (b *BotInfo) GetUserID() (value int64) {
	if b == nil {
		return
	}
	return b.UserID
}

// GetDescription returns value of Description field.
func (b *BotInfo) GetDescription() (value string) {
	if b == nil {
		return
	}
	return b.Description
}

// GetCommands returns value of Commands field.
func (b *BotInfo) GetCommands() (value []BotCommand) {
	if b == nil {
		return
	}
	return b.Commands
}

// GetMenuButton returns value of MenuButton field.
func (b *BotInfo) GetMenuButton() (value BotMenuButtonClass) {
	if b == nil {
		return
	}
	return b.MenuButton
}
