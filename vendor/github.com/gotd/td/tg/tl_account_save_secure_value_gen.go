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

// AccountSaveSecureValueRequest represents TL type `account.saveSecureValue#899fe31d`.
// Securely save Telegram Passport¹ document, for more info see the passport docs »²
//
// Links:
//  1) https://core.telegram.org/passport
//  2) https://core.telegram.org/passport/encryption#encryption
//
// See https://core.telegram.org/method/account.saveSecureValue for reference.
type AccountSaveSecureValueRequest struct {
	// Secure value, for more info see the passport docs »¹
	//
	// Links:
	//  1) https://core.telegram.org/passport/encryption#encryption
	Value InputSecureValue
	// Passport secret hash, for more info see the passport docs »¹
	//
	// Links:
	//  1) https://core.telegram.org/passport/encryption#encryption
	SecureSecretID int64
}

// AccountSaveSecureValueRequestTypeID is TL type id of AccountSaveSecureValueRequest.
const AccountSaveSecureValueRequestTypeID = 0x899fe31d

// Ensuring interfaces in compile-time for AccountSaveSecureValueRequest.
var (
	_ bin.Encoder     = &AccountSaveSecureValueRequest{}
	_ bin.Decoder     = &AccountSaveSecureValueRequest{}
	_ bin.BareEncoder = &AccountSaveSecureValueRequest{}
	_ bin.BareDecoder = &AccountSaveSecureValueRequest{}
)

func (s *AccountSaveSecureValueRequest) Zero() bool {
	if s == nil {
		return true
	}
	if !(s.Value.Zero()) {
		return false
	}
	if !(s.SecureSecretID == 0) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (s *AccountSaveSecureValueRequest) String() string {
	if s == nil {
		return "AccountSaveSecureValueRequest(nil)"
	}
	type Alias AccountSaveSecureValueRequest
	return fmt.Sprintf("AccountSaveSecureValueRequest%+v", Alias(*s))
}

// FillFrom fills AccountSaveSecureValueRequest from given interface.
func (s *AccountSaveSecureValueRequest) FillFrom(from interface {
	GetValue() (value InputSecureValue)
	GetSecureSecretID() (value int64)
}) {
	s.Value = from.GetValue()
	s.SecureSecretID = from.GetSecureSecretID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*AccountSaveSecureValueRequest) TypeID() uint32 {
	return AccountSaveSecureValueRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*AccountSaveSecureValueRequest) TypeName() string {
	return "account.saveSecureValue"
}

// TypeInfo returns info about TL type.
func (s *AccountSaveSecureValueRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "account.saveSecureValue",
		ID:   AccountSaveSecureValueRequestTypeID,
	}
	if s == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Value",
			SchemaName: "value",
		},
		{
			Name:       "SecureSecretID",
			SchemaName: "secure_secret_id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (s *AccountSaveSecureValueRequest) Encode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode account.saveSecureValue#899fe31d as nil")
	}
	b.PutID(AccountSaveSecureValueRequestTypeID)
	return s.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (s *AccountSaveSecureValueRequest) EncodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't encode account.saveSecureValue#899fe31d as nil")
	}
	if err := s.Value.Encode(b); err != nil {
		return fmt.Errorf("unable to encode account.saveSecureValue#899fe31d: field value: %w", err)
	}
	b.PutLong(s.SecureSecretID)
	return nil
}

// Decode implements bin.Decoder.
func (s *AccountSaveSecureValueRequest) Decode(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode account.saveSecureValue#899fe31d to nil")
	}
	if err := b.ConsumeID(AccountSaveSecureValueRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode account.saveSecureValue#899fe31d: %w", err)
	}
	return s.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (s *AccountSaveSecureValueRequest) DecodeBare(b *bin.Buffer) error {
	if s == nil {
		return fmt.Errorf("can't decode account.saveSecureValue#899fe31d to nil")
	}
	{
		if err := s.Value.Decode(b); err != nil {
			return fmt.Errorf("unable to decode account.saveSecureValue#899fe31d: field value: %w", err)
		}
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode account.saveSecureValue#899fe31d: field secure_secret_id: %w", err)
		}
		s.SecureSecretID = value
	}
	return nil
}

// GetValue returns value of Value field.
func (s *AccountSaveSecureValueRequest) GetValue() (value InputSecureValue) {
	if s == nil {
		return
	}
	return s.Value
}

// GetSecureSecretID returns value of SecureSecretID field.
func (s *AccountSaveSecureValueRequest) GetSecureSecretID() (value int64) {
	if s == nil {
		return
	}
	return s.SecureSecretID
}

// AccountSaveSecureValue invokes method account.saveSecureValue#899fe31d returning error if any.
// Securely save Telegram Passport¹ document, for more info see the passport docs »²
//
// Links:
//  1) https://core.telegram.org/passport
//  2) https://core.telegram.org/passport/encryption#encryption
//
// Possible errors:
//  400 PASSWORD_REQUIRED: A 2FA password must be configured to use Telegram Passport.
//
// See https://core.telegram.org/method/account.saveSecureValue for reference.
func (c *Client) AccountSaveSecureValue(ctx context.Context, request *AccountSaveSecureValueRequest) (*SecureValue, error) {
	var result SecureValue

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}