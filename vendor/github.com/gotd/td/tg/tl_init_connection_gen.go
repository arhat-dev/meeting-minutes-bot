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

// InitConnectionRequest represents TL type `initConnection#c1cd5ea9`.
// Initialize connection
//
// See https://core.telegram.org/constructor/initConnection for reference.
type InitConnectionRequest struct {
	// Flags, see TL conditional fields¹
	//
	// Links:
	//  1) https://core.telegram.org/mtproto/TL-combinators#conditional-fields
	Flags bin.Fields
	// Application identifier (see. App configuration¹)
	//
	// Links:
	//  1) https://core.telegram.org/myapp
	APIID int
	// Device model
	DeviceModel string
	// Operation system version
	SystemVersion string
	// Application version
	AppVersion string
	// Code for the language used on the device's OS, ISO 639-1 standard
	SystemLangCode string
	// Language pack to use
	LangPack string
	// Code for the language used on the client, ISO 639-1 standard
	LangCode string
	// Info about an MTProto proxy
	//
	// Use SetProxy and GetProxy helpers.
	Proxy InputClientProxy
	// Additional initConnection parameters. For now, only the tz_offset field is supported,
	// for specifying timezone offset in seconds.
	//
	// Use SetParams and GetParams helpers.
	Params JSONValueClass
	// The query itself
	Query bin.Object
}

// InitConnectionRequestTypeID is TL type id of InitConnectionRequest.
const InitConnectionRequestTypeID = 0xc1cd5ea9

// Ensuring interfaces in compile-time for InitConnectionRequest.
var (
	_ bin.Encoder     = &InitConnectionRequest{}
	_ bin.Decoder     = &InitConnectionRequest{}
	_ bin.BareEncoder = &InitConnectionRequest{}
	_ bin.BareDecoder = &InitConnectionRequest{}
)

func (i *InitConnectionRequest) Zero() bool {
	if i == nil {
		return true
	}
	if !(i.Flags.Zero()) {
		return false
	}
	if !(i.APIID == 0) {
		return false
	}
	if !(i.DeviceModel == "") {
		return false
	}
	if !(i.SystemVersion == "") {
		return false
	}
	if !(i.AppVersion == "") {
		return false
	}
	if !(i.SystemLangCode == "") {
		return false
	}
	if !(i.LangPack == "") {
		return false
	}
	if !(i.LangCode == "") {
		return false
	}
	if !(i.Proxy.Zero()) {
		return false
	}
	if !(i.Params == nil) {
		return false
	}
	if !(i.Query == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (i *InitConnectionRequest) String() string {
	if i == nil {
		return "InitConnectionRequest(nil)"
	}
	type Alias InitConnectionRequest
	return fmt.Sprintf("InitConnectionRequest%+v", Alias(*i))
}

// FillFrom fills InitConnectionRequest from given interface.
func (i *InitConnectionRequest) FillFrom(from interface {
	GetAPIID() (value int)
	GetDeviceModel() (value string)
	GetSystemVersion() (value string)
	GetAppVersion() (value string)
	GetSystemLangCode() (value string)
	GetLangPack() (value string)
	GetLangCode() (value string)
	GetProxy() (value InputClientProxy, ok bool)
	GetParams() (value JSONValueClass, ok bool)
	GetQuery() (value bin.Object)
}) {
	i.APIID = from.GetAPIID()
	i.DeviceModel = from.GetDeviceModel()
	i.SystemVersion = from.GetSystemVersion()
	i.AppVersion = from.GetAppVersion()
	i.SystemLangCode = from.GetSystemLangCode()
	i.LangPack = from.GetLangPack()
	i.LangCode = from.GetLangCode()
	if val, ok := from.GetProxy(); ok {
		i.Proxy = val
	}

	if val, ok := from.GetParams(); ok {
		i.Params = val
	}

	i.Query = from.GetQuery()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*InitConnectionRequest) TypeID() uint32 {
	return InitConnectionRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*InitConnectionRequest) TypeName() string {
	return "initConnection"
}

// TypeInfo returns info about TL type.
func (i *InitConnectionRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "initConnection",
		ID:   InitConnectionRequestTypeID,
	}
	if i == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "APIID",
			SchemaName: "api_id",
		},
		{
			Name:       "DeviceModel",
			SchemaName: "device_model",
		},
		{
			Name:       "SystemVersion",
			SchemaName: "system_version",
		},
		{
			Name:       "AppVersion",
			SchemaName: "app_version",
		},
		{
			Name:       "SystemLangCode",
			SchemaName: "system_lang_code",
		},
		{
			Name:       "LangPack",
			SchemaName: "lang_pack",
		},
		{
			Name:       "LangCode",
			SchemaName: "lang_code",
		},
		{
			Name:       "Proxy",
			SchemaName: "proxy",
			Null:       !i.Flags.Has(0),
		},
		{
			Name:       "Params",
			SchemaName: "params",
			Null:       !i.Flags.Has(1),
		},
		{
			Name:       "Query",
			SchemaName: "query",
		},
	}
	return typ
}

// SetFlags sets flags for non-zero fields.
func (i *InitConnectionRequest) SetFlags() {
	if !(i.Proxy.Zero()) {
		i.Flags.Set(0)
	}
	if !(i.Params == nil) {
		i.Flags.Set(1)
	}
}

// Encode implements bin.Encoder.
func (i *InitConnectionRequest) Encode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode initConnection#c1cd5ea9 as nil")
	}
	b.PutID(InitConnectionRequestTypeID)
	return i.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (i *InitConnectionRequest) EncodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't encode initConnection#c1cd5ea9 as nil")
	}
	i.SetFlags()
	if err := i.Flags.Encode(b); err != nil {
		return fmt.Errorf("unable to encode initConnection#c1cd5ea9: field flags: %w", err)
	}
	b.PutInt(i.APIID)
	b.PutString(i.DeviceModel)
	b.PutString(i.SystemVersion)
	b.PutString(i.AppVersion)
	b.PutString(i.SystemLangCode)
	b.PutString(i.LangPack)
	b.PutString(i.LangCode)
	if i.Flags.Has(0) {
		if err := i.Proxy.Encode(b); err != nil {
			return fmt.Errorf("unable to encode initConnection#c1cd5ea9: field proxy: %w", err)
		}
	}
	if i.Flags.Has(1) {
		if i.Params == nil {
			return fmt.Errorf("unable to encode initConnection#c1cd5ea9: field params is nil")
		}
		if err := i.Params.Encode(b); err != nil {
			return fmt.Errorf("unable to encode initConnection#c1cd5ea9: field params: %w", err)
		}
	}
	if err := i.Query.Encode(b); err != nil {
		return fmt.Errorf("unable to encode initConnection#c1cd5ea9: field query: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (i *InitConnectionRequest) Decode(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode initConnection#c1cd5ea9 to nil")
	}
	if err := b.ConsumeID(InitConnectionRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode initConnection#c1cd5ea9: %w", err)
	}
	return i.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (i *InitConnectionRequest) DecodeBare(b *bin.Buffer) error {
	if i == nil {
		return fmt.Errorf("can't decode initConnection#c1cd5ea9 to nil")
	}
	{
		if err := i.Flags.Decode(b); err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field flags: %w", err)
		}
	}
	{
		value, err := b.Int()
		if err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field api_id: %w", err)
		}
		i.APIID = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field device_model: %w", err)
		}
		i.DeviceModel = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field system_version: %w", err)
		}
		i.SystemVersion = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field app_version: %w", err)
		}
		i.AppVersion = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field system_lang_code: %w", err)
		}
		i.SystemLangCode = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field lang_pack: %w", err)
		}
		i.LangPack = value
	}
	{
		value, err := b.String()
		if err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field lang_code: %w", err)
		}
		i.LangCode = value
	}
	if i.Flags.Has(0) {
		if err := i.Proxy.Decode(b); err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field proxy: %w", err)
		}
	}
	if i.Flags.Has(1) {
		value, err := DecodeJSONValue(b)
		if err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field params: %w", err)
		}
		i.Params = value
	}
	{
		if err := i.Query.Decode(b); err != nil {
			return fmt.Errorf("unable to decode initConnection#c1cd5ea9: field query: %w", err)
		}
	}
	return nil
}

// GetAPIID returns value of APIID field.
func (i *InitConnectionRequest) GetAPIID() (value int) {
	if i == nil {
		return
	}
	return i.APIID
}

// GetDeviceModel returns value of DeviceModel field.
func (i *InitConnectionRequest) GetDeviceModel() (value string) {
	if i == nil {
		return
	}
	return i.DeviceModel
}

// GetSystemVersion returns value of SystemVersion field.
func (i *InitConnectionRequest) GetSystemVersion() (value string) {
	if i == nil {
		return
	}
	return i.SystemVersion
}

// GetAppVersion returns value of AppVersion field.
func (i *InitConnectionRequest) GetAppVersion() (value string) {
	if i == nil {
		return
	}
	return i.AppVersion
}

// GetSystemLangCode returns value of SystemLangCode field.
func (i *InitConnectionRequest) GetSystemLangCode() (value string) {
	if i == nil {
		return
	}
	return i.SystemLangCode
}

// GetLangPack returns value of LangPack field.
func (i *InitConnectionRequest) GetLangPack() (value string) {
	if i == nil {
		return
	}
	return i.LangPack
}

// GetLangCode returns value of LangCode field.
func (i *InitConnectionRequest) GetLangCode() (value string) {
	if i == nil {
		return
	}
	return i.LangCode
}

// SetProxy sets value of Proxy conditional field.
func (i *InitConnectionRequest) SetProxy(value InputClientProxy) {
	i.Flags.Set(0)
	i.Proxy = value
}

// GetProxy returns value of Proxy conditional field and
// boolean which is true if field was set.
func (i *InitConnectionRequest) GetProxy() (value InputClientProxy, ok bool) {
	if i == nil {
		return
	}
	if !i.Flags.Has(0) {
		return value, false
	}
	return i.Proxy, true
}

// SetParams sets value of Params conditional field.
func (i *InitConnectionRequest) SetParams(value JSONValueClass) {
	i.Flags.Set(1)
	i.Params = value
}

// GetParams returns value of Params conditional field and
// boolean which is true if field was set.
func (i *InitConnectionRequest) GetParams() (value JSONValueClass, ok bool) {
	if i == nil {
		return
	}
	if !i.Flags.Has(1) {
		return value, false
	}
	return i.Params, true
}

// GetQuery returns value of Query field.
func (i *InitConnectionRequest) GetQuery() (value bin.Object) {
	if i == nil {
		return
	}
	return i.Query
}