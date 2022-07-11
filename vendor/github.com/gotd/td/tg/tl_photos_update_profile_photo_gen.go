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

// PhotosUpdateProfilePhotoRequest represents TL type `photos.updateProfilePhoto#72d4742c`.
// Installs a previously uploaded photo as a profile photo.
//
// See https://core.telegram.org/method/photos.updateProfilePhoto for reference.
type PhotosUpdateProfilePhotoRequest struct {
	// Input photo
	ID InputPhotoClass
}

// PhotosUpdateProfilePhotoRequestTypeID is TL type id of PhotosUpdateProfilePhotoRequest.
const PhotosUpdateProfilePhotoRequestTypeID = 0x72d4742c

// Ensuring interfaces in compile-time for PhotosUpdateProfilePhotoRequest.
var (
	_ bin.Encoder     = &PhotosUpdateProfilePhotoRequest{}
	_ bin.Decoder     = &PhotosUpdateProfilePhotoRequest{}
	_ bin.BareEncoder = &PhotosUpdateProfilePhotoRequest{}
	_ bin.BareDecoder = &PhotosUpdateProfilePhotoRequest{}
)

func (u *PhotosUpdateProfilePhotoRequest) Zero() bool {
	if u == nil {
		return true
	}
	if !(u.ID == nil) {
		return false
	}

	return true
}

// String implements fmt.Stringer.
func (u *PhotosUpdateProfilePhotoRequest) String() string {
	if u == nil {
		return "PhotosUpdateProfilePhotoRequest(nil)"
	}
	type Alias PhotosUpdateProfilePhotoRequest
	return fmt.Sprintf("PhotosUpdateProfilePhotoRequest%+v", Alias(*u))
}

// FillFrom fills PhotosUpdateProfilePhotoRequest from given interface.
func (u *PhotosUpdateProfilePhotoRequest) FillFrom(from interface {
	GetID() (value InputPhotoClass)
}) {
	u.ID = from.GetID()
}

// TypeID returns type id in TL schema.
//
// See https://core.telegram.org/mtproto/TL-tl#remarks.
func (*PhotosUpdateProfilePhotoRequest) TypeID() uint32 {
	return PhotosUpdateProfilePhotoRequestTypeID
}

// TypeName returns name of type in TL schema.
func (*PhotosUpdateProfilePhotoRequest) TypeName() string {
	return "photos.updateProfilePhoto"
}

// TypeInfo returns info about TL type.
func (u *PhotosUpdateProfilePhotoRequest) TypeInfo() tdp.Type {
	typ := tdp.Type{
		Name: "photos.updateProfilePhoto",
		ID:   PhotosUpdateProfilePhotoRequestTypeID,
	}
	if u == nil {
		typ.Null = true
		return typ
	}
	typ.Fields = []tdp.Field{
		{
			Name:       "ID",
			SchemaName: "id",
		},
	}
	return typ
}

// Encode implements bin.Encoder.
func (u *PhotosUpdateProfilePhotoRequest) Encode(b *bin.Buffer) error {
	if u == nil {
		return fmt.Errorf("can't encode photos.updateProfilePhoto#72d4742c as nil")
	}
	b.PutID(PhotosUpdateProfilePhotoRequestTypeID)
	return u.EncodeBare(b)
}

// EncodeBare implements bin.BareEncoder.
func (u *PhotosUpdateProfilePhotoRequest) EncodeBare(b *bin.Buffer) error {
	if u == nil {
		return fmt.Errorf("can't encode photos.updateProfilePhoto#72d4742c as nil")
	}
	if u.ID == nil {
		return fmt.Errorf("unable to encode photos.updateProfilePhoto#72d4742c: field id is nil")
	}
	if err := u.ID.Encode(b); err != nil {
		return fmt.Errorf("unable to encode photos.updateProfilePhoto#72d4742c: field id: %w", err)
	}
	return nil
}

// Decode implements bin.Decoder.
func (u *PhotosUpdateProfilePhotoRequest) Decode(b *bin.Buffer) error {
	if u == nil {
		return fmt.Errorf("can't decode photos.updateProfilePhoto#72d4742c to nil")
	}
	if err := b.ConsumeID(PhotosUpdateProfilePhotoRequestTypeID); err != nil {
		return fmt.Errorf("unable to decode photos.updateProfilePhoto#72d4742c: %w", err)
	}
	return u.DecodeBare(b)
}

// DecodeBare implements bin.BareDecoder.
func (u *PhotosUpdateProfilePhotoRequest) DecodeBare(b *bin.Buffer) error {
	if u == nil {
		return fmt.Errorf("can't decode photos.updateProfilePhoto#72d4742c to nil")
	}
	{
		value, err := DecodeInputPhoto(b)
		if err != nil {
			return fmt.Errorf("unable to decode photos.updateProfilePhoto#72d4742c: field id: %w", err)
		}
		u.ID = value
	}
	return nil
}

// GetID returns value of ID field.
func (u *PhotosUpdateProfilePhotoRequest) GetID() (value InputPhotoClass) {
	if u == nil {
		return
	}
	return u.ID
}

// GetIDAsNotEmpty returns mapped value of ID field.
func (u *PhotosUpdateProfilePhotoRequest) GetIDAsNotEmpty() (*InputPhoto, bool) {
	return u.ID.AsNotEmpty()
}

// PhotosUpdateProfilePhoto invokes method photos.updateProfilePhoto#72d4742c returning error if any.
// Installs a previously uploaded photo as a profile photo.
//
// Possible errors:
//  400 ALBUM_PHOTOS_TOO_MANY: Too many.
//  400 FILE_PARTS_INVALID: The number of file parts is invalid.
//  400 IMAGE_PROCESS_FAILED: Failure while processing image.
//  400 LOCATION_INVALID: The provided location is invalid.
//  400 PHOTO_CROP_SIZE_SMALL: Photo is too small.
//  400 PHOTO_EXT_INVALID: The extension of the photo is invalid.
//  400 PHOTO_ID_INVALID: Photo ID invalid.
//
// See https://core.telegram.org/method/photos.updateProfilePhoto for reference.
func (c *Client) PhotosUpdateProfilePhoto(ctx context.Context, id InputPhotoClass) (*PhotosPhoto, error) {
	var result PhotosPhoto

	request := &PhotosUpdateProfilePhotoRequest{
		ID: id,
	}
	if err := c.rpc.Invoke(ctx, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
