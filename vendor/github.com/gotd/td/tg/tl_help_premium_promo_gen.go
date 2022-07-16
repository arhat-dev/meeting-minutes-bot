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

// HelpPremiumPromo represents TL type `help.premiumPromo#8a4f3c29`.
//
// See https://core.telegram.org/constructor/help.premiumPromo for reference.
type HelpPremiumPromo struct {
	// StatusText field of HelpPremiumPromo.
	StatusText string
	// StatusEntities field of HelpPremiumPromo.
	StatusEntities []MessageEntityClass
	// VideoSections field of HelpPremiumPromo.
	VideoSections []string
	// Videos field of HelpPremiumPromo.
	Videos []DocumentClass
	// Currency field of HelpPremiumPromo.
	Currency string
	// MonthlyAmount field of HelpPremiumPromo.
	MonthlyAmount int64
	// Users field of HelpPremiumPromo.
	Users []UserClass
}

// HelpPremiumPromoTypeID is TL type id of HelpPremiumPromo.
const HelpPremiumPromoTypeID = 0x8a4f3c29

// Ensuring interfaces in compile-time for HelpPremiumPromo.
var (
	_ bin.Encoder     = &HelpPremiumPromo{}
	_ bin.Decoder     = &HelpPremiumPromo{}
	_ bin.BareEncoder = &HelpPremiumPromo{}
	_ bin.BareDecoder = &HelpPremiumPromo{}
)

func (p *HelpPremiumPromo) Zero() bool {
	if p == nil {
		return true
	}
	if !(p.StatusText == "") {
		return false
	}
	if !(p.StatusEntities == nil) {
		return false
	}
	if !(p.VideoSections == nil) {
		return false
	}
	if !(p.Videos == nil) {
		return false
	}
	if !(p.Currency == "") {
		return false
	}
	if !(p.MonthlyAmount == 0) {
		return false
	}
	if !(p.Users == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (p *HelpPremiumPromo) String() string {
	if p == nil {
		return "HelpPremiumPromo(nil)"
	}
	type Alias HelpPremiumPromo
	return fmt.Sprintf("HelpPremiumPromo%+v", Alias(*p))
}

// FillFrom fills HelpPremiumPromo from given interface.
func (p *HelpPremiumPromo) FillFrom(from interface {
	GetStatusText() (value string)
	GetStatusEntities() (value []MessageEntityClass)
	GetVideoSections() (value []string)
	GetVideos() (value []DocumentClass)
	GetCurrency() (value string)
	GetMonthlyAmount() (value int64)
	GetUsers() (value []UserClass)
}) {
	p.StatusText = from.GetStatusText()
	p.StatusEntities = from.GetStatusEntities()
	p.VideoSections = from.GetVideoSections()
	p.Videos = from.GetVideos()
	p.Currency = from.GetCurrency()
	p.MonthlyAmount = from.GetMonthlyAmount()
	p.Users = from.GetUsers()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*HelpPremiumPromo) TypeID() uint32 {
	return HelpPremiumPromoTypeID
}

// TypeName returns name of type in TL schema.
func (*HelpPremiumPromo) TypeName() string {
	return "help.premiumPromo"
}

// TypeInfo returns info about TL type.
func (p *HelpPremiumPromo) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "help.premiumPromo",
		ID:   HelpPremiumPromoTypeID,
	}
	if p == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "StatusText",
			SchemaName: "status_text",
		},
		{
			Name:       "StatusEntities",
			SchemaName: "status_entities",
		},
		{
			Name:       "VideoSections",
			SchemaName: "video_sections",
		},
		{
			Name:       "Videos",
			SchemaName: "videos",
		},
		{
			Name:       "Currency",
			SchemaName: "currency",
		},
		{
			Name:       "MonthlyAmount",
			SchemaName: "monthly_amount",
		},
		{
			Name:       "Users",
			SchemaName: "users",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (p *HelpPremiumPromo) Encode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode help.premiumPromo#8a4f3c29 as nil")
	}
	b.PutID(HelpPremiumPromoTypeID)
	return p.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (p *HelpPremiumPromo) EncodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't encode help.premiumPromo#8a4f3c29 as nil")
	}
	b.PutString(p.StatusText)
	b.PutVectorHeader(len(p.StatusEntities))
	for idx, v := range p.StatusEntities {
		if v == nil {
			return fmt.Errorf("unable to encode help.premiumPromo#8a4f3c29: field status_entities element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode help.premiumPromo#8a4f3c29: field status_entities element with index %d: %w", idx, err)
		}
	}
	b.PutVectorHeader(len(p.VideoSections))
	for _, v := range p.VideoSections {
		b.PutString(v)
	}
	b.PutVectorHeader(len(p.Videos))
	for idx, v := range p.Videos {
		if v == nil {
			return fmt.Errorf("unable to encode help.premiumPromo#8a4f3c29: field videos element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode help.premiumPromo#8a4f3c29: field videos element with index %d: %w", idx, err)
		}
	}
	b.PutString(p.Currency)
	b.PutLong(p.MonthlyAmount)
	b.PutVectorHeader(len(p.Users))
	for idx, v := range p.Users {
		if v == nil {
			return fmt.Errorf("unable to encode help.premiumPromo#8a4f3c29: field users element with index %d is nil", idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode help.premiumPromo#8a4f3c29: field users element with index %d: %w", idx, err)
		}
	}
	return nil
}

// Decode implements bin.Decoder.
func (p *HelpPremiumPromo) Decode(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode help.premiumPromo#8a4f3c29 to nil")
	}
	if err := b.ConsumeID(HelpPremiumPromoTypeID); err != nil {
		return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: %w", err)
	}
	return p.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (p *HelpPremiumPromo) DecodeBare(b *bin.Buffer) error {
	if p == nil {
		return fmt.Errorf("can't decode help.premiumPromo#8a4f3c29 to nil")
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field status_text: %w", err)
		}
		p.StatusText = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field status_entities: %w", err)
		}

		if headerLen > 0 {
			p.StatusEntities = make([]MessageEntityClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeMessageEntity(b)
			if err != nil {
				return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field status_entities: %w", err)
			}
			p.StatusEntities = append(p.StatusEntities, value)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field video_sections: %w", err)
		}

		if headerLen > 0 {
			p.VideoSections = make([]string, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := b.String()
			if err != nil {
				return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field video_sections: %w", err)
			}
			p.VideoSections = append(p.VideoSections, value)
		}
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field videos: %w", err)
		}

		if headerLen > 0 {
			p.Videos = make([]DocumentClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeDocument(b)
			if err != nil {
				return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field videos: %w", err)
			}
			p.Videos = append(p.Videos, value)
		}
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field currency: %w", err)
		}
		p.Currency = value
	}
	{
		value, err := b.Long()
		if err != nil {
			return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field monthly_amount: %w", err)
		}
		p.MonthlyAmount = value
	}
	{
		headerLen, err := b.VectorHeader()
		if err != nil {
			return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field users: %w", err)
		}

		if headerLen > 0 {
			p.Users = make([]UserClass, 0, headerLen%bin.PreallocateLimit)
		}
		for idx := 0; idx < headerLen; idx++ {
			value, err := DecodeUser(b)
			if err != nil {
				return fmt.Errorf("unable to decode help.premiumPromo#8a4f3c29: field users: %w", err)
			}
			p.Users = append(p.Users, value)
		}
	}
	return nil
}

// GetStatusText returns value of StatusText field.
func (p *HelpPremiumPromo) GetStatusText() (value string) {
	if p == nil {
		return
	}
	return p.StatusText
}

// GetStatusEntities returns value of StatusEntities field.
func (p *HelpPremiumPromo) GetStatusEntities() (value []MessageEntityClass) {
	if p == nil {
		return
	}
	return p.StatusEntities
}

// GetVideoSections returns value of VideoSections field.
func (p *HelpPremiumPromo) GetVideoSections() (value []string) {
	if p == nil {
		return
	}
	return p.VideoSections
}

// GetVideos returns value of Videos field.
func (p *HelpPremiumPromo) GetVideos() (value []DocumentClass) {
	if p == nil {
		return
	}
	return p.Videos
}

// GetCurrency returns value of Currency field.
func (p *HelpPremiumPromo) GetCurrency() (value string) {
	if p == nil {
		return
	}
	return p.Currency
}

// GetMonthlyAmount returns value of MonthlyAmount field.
func (p *HelpPremiumPromo) GetMonthlyAmount() (value int64) {
	if p == nil {
		return
	}
	return p.MonthlyAmount
}

// GetUsers returns value of Users field.
func (p *HelpPremiumPromo) GetUsers() (value []UserClass) {
	if p == nil {
		return
	}
	return p.Users
}

// MapStatusEntities returns field StatusEntities wrapped in MessageEntityClassArray helper.
func (p *HelpPremiumPromo) MapStatusEntities() (value MessageEntityClassArray) {
	return MessageEntityClassArray(p.StatusEntities)
}

// MapVideos returns field Videos wrapped in DocumentClassArray helper.
func (p *HelpPremiumPromo) MapVideos() (value DocumentClassArray) {
	return DocumentClassArray(p.Videos)
}

// MapUsers returns field Users wrapped in UserClassArray helper.
func (p *HelpPremiumPromo) MapUsers() (value UserClassArray) {
	return UserClassArray(p.Users)
}
