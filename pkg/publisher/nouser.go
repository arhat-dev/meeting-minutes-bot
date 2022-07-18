package publisher

import "arhat.dev/mbot/pkg/rt"

var _ User = NoUser{}

type NoUser struct{}

func (NoUser) NextExepcted() rt.LoginFlow { return rt.LoginFlow_None }
func (NoUser) SetPassword(string)         {}
func (NoUser) SetTOTPCode(string)         {}
func (NoUser) SetToken(string)            {}
func (NoUser) SetUsername(string)         {}
