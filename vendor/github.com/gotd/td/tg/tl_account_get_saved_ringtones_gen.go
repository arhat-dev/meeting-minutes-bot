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

// AccountGetSavedRingtonesRequest represents TL type `account.getSavedRingtones#e1902288`.
//
// See https://core.telegram.org/method/account.getSavedRingtones for reference.
type AccountGetSavedRingtonesRequest struct {
	// Hash field of AccountGetSavedRingtonesRequest.
	Hash int64
}

// AccountGetSavedRingtonesRequestTypeID is TL type id of AccountGetSavedRingtonesRequest.
const AccountGetSavedRingtonesRequestTypeID = 0xe1902288

// Ensuring interfaces in compile-time for AccountGetSavedRingtonesRequest.
var (
	_ bin.Encoder     = &AccountGetSavedRingtonesRequest{}
	_ bin.Decoder     = &AccountGetSavedRingtonesRequest{}
	_ bin.BareEncoder = &AccountGetSavedRingtonesRequest{}
	_ bin.BareDecoder = &AccountGetSavedRingtonesRequest{}
)

func (g *AccountGetSavedRingtonesRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Hash == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *AccountGetSavedRingtonesRequest) String() string {
	if g == nil {
		return "AccountGetSavedRingtonesRequest(nil)"
	}
	type Alias AccountGetSavedRingtonesRequest
	return fmt.Sprintf("AccountGetSavedRingtonesRequest%+v", Alias(*g))
}

// FillFrom fills AccountGetSavedRingtonesRequest from given interface.
func (g *AccountGetSavedRingtonesRequest) FillFrom(from interface {
	GetHash() (value int64)
}) {
	g.Hash = from.GetHash()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*AccountGetSavedRingtonesRequest) TypeID() uint32 {
	return AccountGetSavedRingtonesRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*AccountGetSavedRingtonesRequest) TypeName() string {
	return "account.getSavedRingtones"
}

// TypeInfo returns info about TL type.
func (g *AccountGetSavedRingtonesRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "account.getSavedRingtones",
		ID:   AccountGetSavedRingtonesRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Hash",
			SchemaName: "hash",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (g *AccountGetSavedRingtonesRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode account.getSavedRingtones#e1902288 as nil")
	}
	b.PutID(AccountGetSavedRingtonesRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *AccountGetSavedRingtonesRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode account.getSavedRingtones#e1902288 as nil")
	}
	b.PutLong(g.Hash)
	return nil
}

// Decode implements bin.Decoder.
func (g *AccountGetSavedRingtonesRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode account.getSavedRingtones#e1902288 to nil")
	}
	if err := b.ConsumeID(AccountGetSavedRingtonesRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode account.getSavedRingtones#e1902288: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *AccountGetSavedRingtonesRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode account.getSavedRingtones#e1902288 to nil")
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode account.getSavedRingtones#e1902288: field hash: %w", err)
		}
		g.Hash = value
	}
	return nil
}

// GetHash returns value of Hash field.
func (g *AccountGetSavedRingtonesRequest) GetHash() (value int64) {
	if g == nil {
		return
	}
	return g.Hash
}

// AccountGetSavedRingtones invokes method account.getSavedRingtones#e1902288 returning error if any.
//
// See https://core.telegram.org/method/account.getSavedRingtones for reference.
func (c *Client) AccountGetSavedRingtones(ctx context.Context, hash int64) (AccountSavedRingtonesClass, error) {
	var result AccountSavedRingtonesBox

	request := &AccountGetSavedRingtonesRequest{
		Hash: hash,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return result.SavedRingtones, nil
}
