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

// UploadFileClassArray is adapter for slice of UploadFileClass.
type UploadFileClassArray []UploadFileClass

// Sort sorts slice of UploadFileClass.
func (s UploadFileClassArray) Sort(less func(a, b UploadFileClass) bool) UploadFileClassArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UploadFileClass.
func (s UploadFileClassArray) SortStable(less func(a, b UploadFileClass) bool) UploadFileClassArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UploadFileClass.
func (s UploadFileClassArray) Retain(keep func(x UploadFileClass) bool) UploadFileClassArray {
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
func (s UploadFileClassArray) First() (v UploadFileClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UploadFileClassArray) Last() (v UploadFileClass, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UploadFileClassArray) PopFirst() (v UploadFileClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UploadFileClass
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UploadFileClassArray) Pop() (v UploadFileClass, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// AsUploadFile returns copy with only UploadFile constructors.
func (s UploadFileClassArray) AsUploadFile() (to UploadFileArray) {
	for _, elem := range s {
		value, ok := elem.(*UploadFile)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// AsUploadFileCDNRedirect returns copy with only UploadFileCDNRedirect constructors.
func (s UploadFileClassArray) AsUploadFileCDNRedirect() (to UploadFileCDNRedirectArray) {
	for _, elem := range s {
		value, ok := elem.(*UploadFileCDNRedirect)
		if !ok {
			continue
		}
		to = append(to, *value)
	}

	return to
}

// UploadFileArray is adapter for slice of UploadFile.
type UploadFileArray []UploadFile

// Sort sorts slice of UploadFile.
func (s UploadFileArray) Sort(less func(a, b UploadFile) bool) UploadFileArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UploadFile.
func (s UploadFileArray) SortStable(less func(a, b UploadFile) bool) UploadFileArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UploadFile.
func (s UploadFileArray) Retain(keep func(x UploadFile) bool) UploadFileArray {
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
func (s UploadFileArray) First() (v UploadFile, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UploadFileArray) Last() (v UploadFile, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UploadFileArray) PopFirst() (v UploadFile, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UploadFile
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UploadFileArray) Pop() (v UploadFile, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// UploadFileCDNRedirectArray is adapter for slice of UploadFileCDNRedirect.
type UploadFileCDNRedirectArray []UploadFileCDNRedirect

// Sort sorts slice of UploadFileCDNRedirect.
func (s UploadFileCDNRedirectArray) Sort(less func(a, b UploadFileCDNRedirect) bool) UploadFileCDNRedirectArray {
	sort.Slice(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// SortStable sorts slice of UploadFileCDNRedirect.
func (s UploadFileCDNRedirectArray) SortStable(less func(a, b UploadFileCDNRedirect) bool) UploadFileCDNRedirectArray {
	sort.SliceStable(s, func(i, j int) bool {
		return less(s[i], s[j])
	})
	return s
}

// Retain filters in-place slice of UploadFileCDNRedirect.
func (s UploadFileCDNRedirectArray) Retain(keep func(x UploadFileCDNRedirect) bool) UploadFileCDNRedirectArray {
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
func (s UploadFileCDNRedirectArray) First() (v UploadFileCDNRedirect, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[0], true
}

// Last returns last element of slice (if exists).
func (s UploadFileCDNRedirectArray) Last() (v UploadFileCDNRedirect, ok bool) {
	if len(s) < 1 {
		return
	}
	return s[len(s)-1], true
}

// PopFirst returns first element of slice (if exists) and deletes it.
func (s *UploadFileCDNRedirectArray) PopFirst() (v UploadFileCDNRedirect, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[0]

	// Delete by index from SliceTricks.
	copy(a[0:], a[1:])
	var zero UploadFileCDNRedirect
	a[len(a)-1] = zero
	a = a[:len(a)-1]
	*s = a

	return v, true
}

// Pop returns last element of slice (if exists) and deletes it.
func (s *UploadFileCDNRedirectArray) Pop() (v UploadFileCDNRedirect, ok bool) {
	if s == nil || len(*s) < 1 {
		return
	}

	a := *s
	v = a[len(a)-1]
	a = a[:len(a)-1]
	*s = a

	return v, true
}