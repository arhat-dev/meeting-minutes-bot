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

// PaymentsGetPaymentFormRequest represents TL type `payments.getPaymentForm#37148dbb`.
// Get a payment form
//
// See https://core.telegram.org/method/payments.getPaymentForm for reference.
type PaymentsGetPaymentFormRequest struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Invoice field of PaymentsGetPaymentFormRequest.
	Invoice InputInvoiceClass
	// A JSON object with the following keys, containing color theme information (integers,
	// RGB24) to pass to the payment provider, to apply in eventual verification pages:
	// bg_color - Background color text_color - Text color hint_color - Hint text color
	// link_color - Link color button_color - Button color button_text_color - Button text
	// color
	//
	// Use SetThemeParams and GetThemeParams helpers.
	ThemeParams DataJSON
}

// PaymentsGetPaymentFormRequestTypeID is TL type id of PaymentsGetPaymentFormRequest.
const PaymentsGetPaymentFormRequestTypeID = 0x37148dbb

// Ensuring interfaces in compile-time for PaymentsGetPaymentFormRequest.
var (
	_ bin.Encoder     = &PaymentsGetPaymentFormRequest{}
	_ bin.Decoder     = &PaymentsGetPaymentFormRequest{}
	_ bin.BareEncoder = &PaymentsGetPaymentFormRequest{}
	_ bin.BareDecoder = &PaymentsGetPaymentFormRequest{}
)

func (g *PaymentsGetPaymentFormRequest) Zero() bool {
	if g == nil {
		return true
	}
	if !(g.Flags.Zero()) {
		return false
	}
	if !(g.Invoice == nil) {
		return false
	}
	if !(g.ThemeParams.Zero()) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (g *PaymentsGetPaymentFormRequest) String() string {
	if g == nil {
		return "PaymentsGetPaymentFormRequest(nil)"
	}
	type Alias PaymentsGetPaymentFormRequest
	return fmt.Sprintf("PaymentsGetPaymentFormRequest%+v", Alias(*g))
}

// FillFrom fills PaymentsGetPaymentFormRequest from given interface.
func (g *PaymentsGetPaymentFormRequest) FillFrom(from interface {
	GetInvoice() (value InputInvoiceClass)
	GetThemeParams() (value DataJSON, ok bool)
}) {
	g.Invoice = from.GetInvoice()
	if val, ok := from.GetThemeParams(); ok {
		g.ThemeParams = val
	}

}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PaymentsGetPaymentFormRequest) TypeID() uint32 {
	return PaymentsGetPaymentFormRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PaymentsGetPaymentFormRequest) TypeName() string {
	return "payments.getPaymentForm"
}

// TypeInfo returns info about TL type.
func (g *PaymentsGetPaymentFormRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "payments.getPaymentForm",
		ID:   PaymentsGetPaymentFormRequestTypeID,
	}
	if g == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Invoice",
			SchemaName: "invoice",
		},
		{
			Name:       "ThemeParams",
			SchemaName: "theme_params",
			Null:       !g.Flags.Has(0),
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (g *PaymentsGetPaymentFormRequest) SetFlags() {
	if !(g.ThemeParams.Zero()) {
		g.Flags.Set(0)
	}
}

// Encode implements bin.Encoder.
func (g *PaymentsGetPaymentFormRequest) Encode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode payments.getPaymentForm#37148dbb as nil")
	}
	b.PutID(PaymentsGetPaymentFormRequestTypeID)
	return g.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (g *PaymentsGetPaymentFormRequest) EncodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't encode payments.getPaymentForm#37148dbb as nil")
	}
	g.SetFlags()
	if err := g.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode payments.getPaymentForm#37148dbb: field flags: %w", err)
	}
	if g.Invoice == nil {
		return fmt.Errorf("unable to encode payments.getPaymentForm#37148dbb: field invoice is nil")
	}
	if err := g.Invoice.Encode(b); err != nil {
		return fmt.Errorf("unable to encode payments.getPaymentForm#37148dbb: field invoice: %w", err)
	}
	if g.Flags.Has(0) {
		if err := g.ThemeParams.Encode(b); err != nil {
			return fmt.Errorf("unable to encode payments.getPaymentForm#37148dbb: field theme_params: %w", err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (g *PaymentsGetPaymentFormRequest) Decode(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode payments.getPaymentForm#37148dbb to nil")
	}
	if err := b.ConsumeID(PaymentsGetPaymentFormRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode payments.getPaymentForm#37148dbb: %w", err)
	}
	return g.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (g *PaymentsGetPaymentFormRequest) DecodeBare(b *bin.Buffer) error {
	if g == nil {
		return fmt.Errorf("can't decode payments.getPaymentForm#37148dbb to nil")
	}
	{
		if err := g.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode payments.getPaymentForm#37148dbb: field flags: %w", err)
		}
	}
	{
		value, err := DecodeInputInvoice(b)
		if err != nil {
			return fmt.Errorf("unable to decode payments.getPaymentForm#37148dbb: field invoice: %w", err)
		}
		g.Invoice = value
	}
	if g.Flags.Has(0) {
		if err := g.ThemeParams.Decode(b); err != nil {
			return fmt.Errorf("unable to decode payments.getPaymentForm#37148dbb: field theme_params: %w", err)
		}
	}
	return nil
}

// GetInvoice returns value of Invoice field.
func (g *PaymentsGetPaymentFormRequest) GetInvoice() (value InputInvoiceClass) {
	if g == nil {
		return
	}
	return g.Invoice
}

// SetThemeParams sets value of ThemeParams conditional field.
func (g *PaymentsGetPaymentFormRequest) SetThemeParams(value DataJSON) {
	g.Flags.Set(0)
	g.ThemeParams = value
}

// GetThemeParams returns value of ThemeParams conditional field and
// boolean which is true if field was set.
func (g *PaymentsGetPaymentFormRequest) GetThemeParams() (value DataJSON, ok bool) {
	if g == nil {
		return
	}
	if !g.Flags.Has(0) {
		return value, false
	}
	return g.ThemeParams, true
}

// PaymentsGetPaymentForm invokes method payments.getPaymentForm#37148dbb returning error if any.
// Get a payment form
//
// Possible errors:
//  400 MESSAGE_ID_INVALID: The provided message id is invalid.
//
// See https://core.telegram.org/method/payments.getPaymentForm for reference.
func (c *Client) PaymentsGetPaymentForm(ctx context.Context, request *PaymentsGetPaymentFormRequest) (*PaymentsPaymentForm, error) {
	var result PaymentsPaymentForm

	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
