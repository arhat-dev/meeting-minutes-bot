package publisher

import "arhat.dev/mbot/pkg/rt"

var _ User = NoUser{}

// NoUser is a User implementation for publishers doesn't need user login
type NoUser struct{}

func (NoUser) NextCredential() rt.LoginFlow { return rt.LoginFlow_None }
func (NoUser) SetToken(string)              {}
func (NoUser) SetUsername(string)           {}
func (NoUser) SetPassword(string)           {}
func (NoUser) SetTOTPCode(string)           {}
