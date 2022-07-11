//go:build !no_gotd_slices
// +build !no_gotd_slices

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

// KeyboardButtonClassArray is adapter for slice of KeyboardButtonClass.
type KeyboardButtonClassArray []KeyboardButtonClass

// Sort sorts slice of KeyboardButtonClass.
func (s KeyboardButtonClassArray) Sort(less func(a, b KeyboardButtonClass) bool) KeyboardButtonClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonClass.
func (s KeyboardButtonClassArray) SortStable(less func(a, b KeyboardButtonClass) bool) KeyboardButtonClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonClass.
func (s KeyboardButtonClassArray) Retain(keep func(x KeyboardButtonClass) bool) KeyboardButtonClassArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonClassArray) First() (v KeyboardButtonClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonClassArray) Last() (v KeyboardButtonClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonClassArray) PopFirst() (v KeyboardButtonClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonClassArray) Pop() (v KeyboardButtonClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsKeyboardButton returns copy with only KeyboardButton constructors.
func (s KeyboardButtonClassArray) AsKeyboardButton() (to KeyboardButtonArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButton)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonURL returns copy with only KeyboardButtonURL constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonURL() (to KeyboardButtonURLArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonURL)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonCallback returns copy with only KeyboardButtonCallback constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonCallback() (to KeyboardButtonCallbackArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonCallback)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonRequestPhone returns copy with only KeyboardButtonRequestPhone constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonRequestPhone() (to KeyboardButtonRequestPhoneArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonRequestPhone)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonRequestGeoLocation returns copy with only KeyboardButtonRequestGeoLocation constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonRequestGeoLocation() (to KeyboardButtonRequestGeoLocationArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonRequestGeoLocation)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonSwitchInline returns copy with only KeyboardButtonSwitchInline constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonSwitchInline() (to KeyboardButtonSwitchInlineArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonSwitchInline)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonGame returns copy with only KeyboardButtonGame constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonGame() (to KeyboardButtonGameArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonGame)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonBuy returns copy with only KeyboardButtonBuy constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonBuy() (to KeyboardButtonBuyArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonBuy)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonURLAuth returns copy with only KeyboardButtonURLAuth constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonURLAuth() (to KeyboardButtonURLAuthArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonURLAuth)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsInputKeyboardButtonURLAuth returns copy with only InputKeyboardButtonURLAuth constructors.
func (s KeyboardButtonClassArray) AsInputKeyboardButtonURLAuth() (to InputKeyboardButtonURLAuthArray) {
	for _, elem := range s {
		value, ok := elem.(*InputKeyboardButtonURLAuth)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonRequestPoll returns copy with only KeyboardButtonRequestPoll constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonRequestPoll() (to KeyboardButtonRequestPollArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonRequestPoll)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsInputKeyboardButtonUserProfile returns copy with only InputKeyboardButtonUserProfile constructors.
func (s KeyboardButtonClassArray) AsInputKeyboardButtonUserProfile() (to InputKeyboardButtonUserProfileArray) {
	for _, elem := range s {
		value, ok := elem.(*InputKeyboardButtonUserProfile)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonUserProfile returns copy with only KeyboardButtonUserProfile constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonUserProfile() (to KeyboardButtonUserProfileArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonUserProfile)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonWebView returns copy with only KeyboardButtonWebView constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonWebView() (to KeyboardButtonWebViewArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonWebView)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsKeyboardButtonSimpleWebView returns copy with only KeyboardButtonSimpleWebView constructors.
func (s KeyboardButtonClassArray) AsKeyboardButtonSimpleWebView() (to KeyboardButtonSimpleWebViewArray) {
	for _, elem := range s {
		value, ok := elem.(*KeyboardButtonSimpleWebView)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// KeyboardButtonArray is adapter for slice of KeyboardButton.
type KeyboardButtonArray []KeyboardButton

// Sort sorts slice of KeyboardButton.
func (s KeyboardButtonArray) Sort(less func(a, b KeyboardButton) bool) KeyboardButtonArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButton.
func (s KeyboardButtonArray) SortStable(less func(a, b KeyboardButton) bool) KeyboardButtonArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButton.
func (s KeyboardButtonArray) Retain(keep func(x KeyboardButton) bool) KeyboardButtonArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonArray) First() (v KeyboardButton, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonArray) Last() (v KeyboardButton, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonArray) PopFirst() (v KeyboardButton, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButton
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonArray) Pop() (v KeyboardButton, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonURLArray is adapter for slice of KeyboardButtonURL.
type KeyboardButtonURLArray []KeyboardButtonURL

// Sort sorts slice of KeyboardButtonURL.
func (s KeyboardButtonURLArray) Sort(less func(a, b KeyboardButtonURL) bool) KeyboardButtonURLArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonURL.
func (s KeyboardButtonURLArray) SortStable(less func(a, b KeyboardButtonURL) bool) KeyboardButtonURLArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonURL.
func (s KeyboardButtonURLArray) Retain(keep func(x KeyboardButtonURL) bool) KeyboardButtonURLArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonURLArray) First() (v KeyboardButtonURL, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonURLArray) Last() (v KeyboardButtonURL, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonURLArray) PopFirst() (v KeyboardButtonURL, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonURL
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonURLArray) Pop() (v KeyboardButtonURL, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonCallbackArray is adapter for slice of KeyboardButtonCallback.
type KeyboardButtonCallbackArray []KeyboardButtonCallback

// Sort sorts slice of KeyboardButtonCallback.
func (s KeyboardButtonCallbackArray) Sort(less func(a, b KeyboardButtonCallback) bool) KeyboardButtonCallbackArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonCallback.
func (s KeyboardButtonCallbackArray) SortStable(less func(a, b KeyboardButtonCallback) bool) KeyboardButtonCallbackArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonCallback.
func (s KeyboardButtonCallbackArray) Retain(keep func(x KeyboardButtonCallback) bool) KeyboardButtonCallbackArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonCallbackArray) First() (v KeyboardButtonCallback, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonCallbackArray) Last() (v KeyboardButtonCallback, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonCallbackArray) PopFirst() (v KeyboardButtonCallback, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonCallback
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonCallbackArray) Pop() (v KeyboardButtonCallback, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonRequestPhoneArray is adapter for slice of KeyboardButtonRequestPhone.
type KeyboardButtonRequestPhoneArray []KeyboardButtonRequestPhone

// Sort sorts slice of KeyboardButtonRequestPhone.
func (s KeyboardButtonRequestPhoneArray) Sort(less func(a, b KeyboardButtonRequestPhone) bool) KeyboardButtonRequestPhoneArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonRequestPhone.
func (s KeyboardButtonRequestPhoneArray) SortStable(less func(a, b KeyboardButtonRequestPhone) bool) KeyboardButtonRequestPhoneArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonRequestPhone.
func (s KeyboardButtonRequestPhoneArray) Retain(keep func(x KeyboardButtonRequestPhone) bool) KeyboardButtonRequestPhoneArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonRequestPhoneArray) First() (v KeyboardButtonRequestPhone, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonRequestPhoneArray) Last() (v KeyboardButtonRequestPhone, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonRequestPhoneArray) PopFirst() (v KeyboardButtonRequestPhone, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonRequestPhone
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonRequestPhoneArray) Pop() (v KeyboardButtonRequestPhone, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonRequestGeoLocationArray is adapter for slice of KeyboardButtonRequestGeoLocation.
type KeyboardButtonRequestGeoLocationArray []KeyboardButtonRequestGeoLocation

// Sort sorts slice of KeyboardButtonRequestGeoLocation.
func (s KeyboardButtonRequestGeoLocationArray) Sort(less func(a, b KeyboardButtonRequestGeoLocation) bool) KeyboardButtonRequestGeoLocationArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonRequestGeoLocation.
func (s KeyboardButtonRequestGeoLocationArray) SortStable(less func(a, b KeyboardButtonRequestGeoLocation) bool) KeyboardButtonRequestGeoLocationArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonRequestGeoLocation.
func (s KeyboardButtonRequestGeoLocationArray) Retain(keep func(x KeyboardButtonRequestGeoLocation) bool) KeyboardButtonRequestGeoLocationArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonRequestGeoLocationArray) First() (v KeyboardButtonRequestGeoLocation, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonRequestGeoLocationArray) Last() (v KeyboardButtonRequestGeoLocation, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonRequestGeoLocationArray) PopFirst() (v KeyboardButtonRequestGeoLocation, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonRequestGeoLocation
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonRequestGeoLocationArray) Pop() (v KeyboardButtonRequestGeoLocation, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonSwitchInlineArray is adapter for slice of KeyboardButtonSwitchInline.
type KeyboardButtonSwitchInlineArray []KeyboardButtonSwitchInline

// Sort sorts slice of KeyboardButtonSwitchInline.
func (s KeyboardButtonSwitchInlineArray) Sort(less func(a, b KeyboardButtonSwitchInline) bool) KeyboardButtonSwitchInlineArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonSwitchInline.
func (s KeyboardButtonSwitchInlineArray) SortStable(less func(a, b KeyboardButtonSwitchInline) bool) KeyboardButtonSwitchInlineArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonSwitchInline.
func (s KeyboardButtonSwitchInlineArray) Retain(keep func(x KeyboardButtonSwitchInline) bool) KeyboardButtonSwitchInlineArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonSwitchInlineArray) First() (v KeyboardButtonSwitchInline, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonSwitchInlineArray) Last() (v KeyboardButtonSwitchInline, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonSwitchInlineArray) PopFirst() (v KeyboardButtonSwitchInline, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonSwitchInline
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonSwitchInlineArray) Pop() (v KeyboardButtonSwitchInline, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonGameArray is adapter for slice of KeyboardButtonGame.
type KeyboardButtonGameArray []KeyboardButtonGame

// Sort sorts slice of KeyboardButtonGame.
func (s KeyboardButtonGameArray) Sort(less func(a, b KeyboardButtonGame) bool) KeyboardButtonGameArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonGame.
func (s KeyboardButtonGameArray) SortStable(less func(a, b KeyboardButtonGame) bool) KeyboardButtonGameArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonGame.
func (s KeyboardButtonGameArray) Retain(keep func(x KeyboardButtonGame) bool) KeyboardButtonGameArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonGameArray) First() (v KeyboardButtonGame, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonGameArray) Last() (v KeyboardButtonGame, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonGameArray) PopFirst() (v KeyboardButtonGame, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonGame
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonGameArray) Pop() (v KeyboardButtonGame, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonBuyArray is adapter for slice of KeyboardButtonBuy.
type KeyboardButtonBuyArray []KeyboardButtonBuy

// Sort sorts slice of KeyboardButtonBuy.
func (s KeyboardButtonBuyArray) Sort(less func(a, b KeyboardButtonBuy) bool) KeyboardButtonBuyArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonBuy.
func (s KeyboardButtonBuyArray) SortStable(less func(a, b KeyboardButtonBuy) bool) KeyboardButtonBuyArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonBuy.
func (s KeyboardButtonBuyArray) Retain(keep func(x KeyboardButtonBuy) bool) KeyboardButtonBuyArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonBuyArray) First() (v KeyboardButtonBuy, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonBuyArray) Last() (v KeyboardButtonBuy, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonBuyArray) PopFirst() (v KeyboardButtonBuy, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonBuy
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonBuyArray) Pop() (v KeyboardButtonBuy, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonURLAuthArray is adapter for slice of KeyboardButtonURLAuth.
type KeyboardButtonURLAuthArray []KeyboardButtonURLAuth

// Sort sorts slice of KeyboardButtonURLAuth.
func (s KeyboardButtonURLAuthArray) Sort(less func(a, b KeyboardButtonURLAuth) bool) KeyboardButtonURLAuthArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonURLAuth.
func (s KeyboardButtonURLAuthArray) SortStable(less func(a, b KeyboardButtonURLAuth) bool) KeyboardButtonURLAuthArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonURLAuth.
func (s KeyboardButtonURLAuthArray) Retain(keep func(x KeyboardButtonURLAuth) bool) KeyboardButtonURLAuthArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonURLAuthArray) First() (v KeyboardButtonURLAuth, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonURLAuthArray) Last() (v KeyboardButtonURLAuth, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonURLAuthArray) PopFirst() (v KeyboardButtonURLAuth, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonURLAuth
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonURLAuthArray) Pop() (v KeyboardButtonURLAuth, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// InputKeyboardButtonURLAuthArray is adapter for slice of InputKeyboardButtonURLAuth.
type InputKeyboardButtonURLAuthArray []InputKeyboardButtonURLAuth

// Sort sorts slice of InputKeyboardButtonURLAuth.
func (s InputKeyboardButtonURLAuthArray) Sort(less func(a, b InputKeyboardButtonURLAuth) bool) InputKeyboardButtonURLAuthArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of InputKeyboardButtonURLAuth.
func (s InputKeyboardButtonURLAuthArray) SortStable(less func(a, b InputKeyboardButtonURLAuth) bool) InputKeyboardButtonURLAuthArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of InputKeyboardButtonURLAuth.
func (s InputKeyboardButtonURLAuthArray) Retain(keep func(x InputKeyboardButtonURLAuth) bool) InputKeyboardButtonURLAuthArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s InputKeyboardButtonURLAuthArray) First() (v InputKeyboardButtonURLAuth, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s InputKeyboardButtonURLAuthArray) Last() (v InputKeyboardButtonURLAuth, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *InputKeyboardButtonURLAuthArray) PopFirst() (v InputKeyboardButtonURLAuth, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero InputKeyboardButtonURLAuth
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *InputKeyboardButtonURLAuthArray) Pop() (v InputKeyboardButtonURLAuth, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonRequestPollArray is adapter for slice of KeyboardButtonRequestPoll.
type KeyboardButtonRequestPollArray []KeyboardButtonRequestPoll

// Sort sorts slice of KeyboardButtonRequestPoll.
func (s KeyboardButtonRequestPollArray) Sort(less func(a, b KeyboardButtonRequestPoll) bool) KeyboardButtonRequestPollArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonRequestPoll.
func (s KeyboardButtonRequestPollArray) SortStable(less func(a, b KeyboardButtonRequestPoll) bool) KeyboardButtonRequestPollArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonRequestPoll.
func (s KeyboardButtonRequestPollArray) Retain(keep func(x KeyboardButtonRequestPoll) bool) KeyboardButtonRequestPollArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonRequestPollArray) First() (v KeyboardButtonRequestPoll, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonRequestPollArray) Last() (v KeyboardButtonRequestPoll, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonRequestPollArray) PopFirst() (v KeyboardButtonRequestPoll, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonRequestPoll
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonRequestPollArray) Pop() (v KeyboardButtonRequestPoll, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// InputKeyboardButtonUserProfileArray is adapter for slice of InputKeyboardButtonUserProfile.
type InputKeyboardButtonUserProfileArray []InputKeyboardButtonUserProfile

// Sort sorts slice of InputKeyboardButtonUserProfile.
func (s InputKeyboardButtonUserProfileArray) Sort(less func(a, b InputKeyboardButtonUserProfile) bool) InputKeyboardButtonUserProfileArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of InputKeyboardButtonUserProfile.
func (s InputKeyboardButtonUserProfileArray) SortStable(less func(a, b InputKeyboardButtonUserProfile) bool) InputKeyboardButtonUserProfileArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of InputKeyboardButtonUserProfile.
func (s InputKeyboardButtonUserProfileArray) Retain(keep func(x InputKeyboardButtonUserProfile) bool) InputKeyboardButtonUserProfileArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s InputKeyboardButtonUserProfileArray) First() (v InputKeyboardButtonUserProfile, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s InputKeyboardButtonUserProfileArray) Last() (v InputKeyboardButtonUserProfile, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *InputKeyboardButtonUserProfileArray) PopFirst() (v InputKeyboardButtonUserProfile, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero InputKeyboardButtonUserProfile
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *InputKeyboardButtonUserProfileArray) Pop() (v InputKeyboardButtonUserProfile, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonUserProfileArray is adapter for slice of KeyboardButtonUserProfile.
type KeyboardButtonUserProfileArray []KeyboardButtonUserProfile

// Sort sorts slice of KeyboardButtonUserProfile.
func (s KeyboardButtonUserProfileArray) Sort(less func(a, b KeyboardButtonUserProfile) bool) KeyboardButtonUserProfileArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonUserProfile.
func (s KeyboardButtonUserProfileArray) SortStable(less func(a, b KeyboardButtonUserProfile) bool) KeyboardButtonUserProfileArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonUserProfile.
func (s KeyboardButtonUserProfileArray) Retain(keep func(x KeyboardButtonUserProfile) bool) KeyboardButtonUserProfileArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonUserProfileArray) First() (v KeyboardButtonUserProfile, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonUserProfileArray) Last() (v KeyboardButtonUserProfile, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonUserProfileArray) PopFirst() (v KeyboardButtonUserProfile, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonUserProfile
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonUserProfileArray) Pop() (v KeyboardButtonUserProfile, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonWebViewArray is adapter for slice of KeyboardButtonWebView.
type KeyboardButtonWebViewArray []KeyboardButtonWebView

// Sort sorts slice of KeyboardButtonWebView.
func (s KeyboardButtonWebViewArray) Sort(less func(a, b KeyboardButtonWebView) bool) KeyboardButtonWebViewArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonWebView.
func (s KeyboardButtonWebViewArray) SortStable(less func(a, b KeyboardButtonWebView) bool) KeyboardButtonWebViewArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonWebView.
func (s KeyboardButtonWebViewArray) Retain(keep func(x KeyboardButtonWebView) bool) KeyboardButtonWebViewArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonWebViewArray) First() (v KeyboardButtonWebView, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonWebViewArray) Last() (v KeyboardButtonWebView, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonWebViewArray) PopFirst() (v KeyboardButtonWebView, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonWebView
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonWebViewArray) Pop() (v KeyboardButtonWebView, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// KeyboardButtonSimpleWebViewArray is adapter for slice of KeyboardButtonSimpleWebView.
type KeyboardButtonSimpleWebViewArray []KeyboardButtonSimpleWebView

// Sort sorts slice of KeyboardButtonSimpleWebView.
func (s KeyboardButtonSimpleWebViewArray) Sort(less func(a, b KeyboardButtonSimpleWebView) bool) KeyboardButtonSimpleWebViewArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of KeyboardButtonSimpleWebView.
func (s KeyboardButtonSimpleWebViewArray) SortStable(less func(a, b KeyboardButtonSimpleWebView) bool) KeyboardButtonSimpleWebViewArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of KeyboardButtonSimpleWebView.
func (s KeyboardButtonSimpleWebViewArray) Retain(keep func(x KeyboardButtonSimpleWebView) bool) KeyboardButtonSimpleWebViewArray {
	n := 0
	for _, x := range s {
		if keep(x) {
			s[n] = x
			n++
		}
	}
	s = s[:n]

	return s
}

// First returns first element of slice (if exists).
func (s KeyboardButtonSimpleWebViewArray) First() (v KeyboardButtonSimpleWebView, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s KeyboardButtonSimpleWebViewArray) Last() (v KeyboardButtonSimpleWebView, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *KeyboardButtonSimpleWebViewArray) PopFirst() (v KeyboardButtonSimpleWebView, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero KeyboardButtonSimpleWebView
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *KeyboardButtonSimpleWebViewArray) Pop() (v KeyboardButtonSimpleWebView, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}
