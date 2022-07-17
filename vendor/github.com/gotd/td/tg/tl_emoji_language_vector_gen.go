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

// EmojiLanguageVector is a box for Vector<EmojiLanguage>
type EmojiLanguageVector struct {
	// Elements of Vector<EmojiLanguage>
	Elems []EmojiLanguage
}

// EmojiLanguageVectorTypeID is TL type id of EmojiLanguageVector.
const EmojiLanguageVectorTypeID = bin.TypeVector

// Ensuring interfaces in compile-time for EmojiLanguageVector.
var (
	_ bin.Encoder     = &EmojiLanguageVector{}
	_ bin.Decoder     = &EmojiLanguageVector{}
	_ bin.BareEncoder = &EmojiLanguageVector{}
	_ bin.BareDecoder = &EmojiLanguageVector{}
)

func (vec *EmojiLanguageVector) Zero() bool {
	if vec == nil {
		return true
	}
	if !(vec.Elems == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (vec *EmojiLanguageVector) String() string {
	if vec == nil {
		return "EmojiLanguageVector(nil)"
	}
	type Alias EmojiLanguageVector
	return fmt.Sprintf("EmojiLanguageVector%+v", Alias(*vec))
}

// FillFrom fills EmojiLanguageVector from given interface.
func (vec *EmojiLanguageVector) FillFrom(from interface {
	GetElems() (value []EmojiLanguage)
}) {
	vec.Elems = from.GetElems()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*EmojiLanguageVector) TypeID() uint32 {
	return EmojiLanguageVectorTypeID
}

// TypeName returns name of type in TL schema.
func (*EmojiLanguageVector) TypeName() string {
	return ""
}

// TypeInfo returns info about TL type.
func (vec *EmojiLanguageVector) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "",
		ID:   EmojiLanguageVectorTypeID,
	}
	if vec == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "Elems",
			SchemaName: "Elems",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (vec *EmojiLanguageVector) Encode(b *bin.Buffer) error {
	if vec == nil {
		return fmt.Errorf("can't encode Vector<EmojiLanguage> as nil")
	}

	return vec.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (vec *EmojiLanguageVector) EncodeBare(b *bin.Buffer) error {
	if vec == nil {
		return fmt.Errorf("can't encode Vector<EmojiLanguage> as nil")
	}
	b.PutVectorHeader(len(vec.Elems))
	for idx, v := range vec.Elems {
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode Vector<EmojiLanguage>: field Elems element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (vec *EmojiLanguageVector) Decode(b *bin.Buffer) error {
	if vec == nil {
		return fmt.Errorf("can't decode Vector<EmojiLanguage> to nil")
	}

	return vec.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (vec *EmojiLanguageVector) DecodeBare(b *bin.Buffer) error {
	if vec == nil {
		return fmt.Errorf("can't decode Vector<EmojiLanguage> to nil")
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode Vector<EmojiLanguage>: field Elems: %w", err)
		}

		if headerLen > 0 {
			vec.Elems = make([]EmojiLanguage, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			var value EmojiLanguage
			if err := value.Decode(b); err != nil {
				return fmt.Errorf("unable to decode Vector<EmojiLanguage>: field Elems: %w", err)
			}
			vec.Elems = append(vec.Elems, value)
		}
	}
	return nil
}

// GetElems returns value of Elems field.
func (vec *EmojiLanguageVector) GetElems() (value []EmojiLanguage) {
	if vec == nil {
		return
	}
	return vec.Elems
}