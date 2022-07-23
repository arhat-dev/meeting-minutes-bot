// Package archiver implements a web archiver for received messages
//
// it extracts all urls and create web archives for them
package archiver

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
	cdp "github.com/chromedp/chromedp"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct{}

// Peek implements generator.Interface
func (d *Driver) Peek(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	cdp.NewContext(con.Context())
	return
}

// New implements generator.Interface
func (d *Driver) New(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return
}

// Continue implements generator.Interface
func (d *Driver) Continue(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return
}

// Generate implements generator.Interface
func (d *Driver) Generate(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return
}
