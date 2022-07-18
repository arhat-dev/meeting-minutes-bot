package multipub

import (
	"fmt"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/rs"
)

const (
	Name = "multigen"
)

func init() {
	publisher.Register(Name, func() publisher.Config { return &Config{} })
}

type Config struct {
	rs.BaseField

	Specs []singleSpec `yaml:"specs"`
}

type pair struct {
	impl publisher.Interface
	user publisher.User
}

// Create implements publisher.Config
func (c *Config) Create() (_ publisher.Interface, _ publisher.User, err error) {
	underlay := make([]pair, len(c.Specs))
	for i := range underlay {
		underlay[i].impl, underlay[i].user, err = c.Specs[i].resolve()
		if err != nil {
			return
		}
	}

	return &Driver{underlay: underlay}, &User{0, underlay}, nil
}

var _ publisher.User = (*User)(nil)

// TODO: implement this user for all underlay user objects
type User struct {
	index    int
	underlay []pair
}

func (u *User) NextExepcted() (flow rt.LoginFlow) {
	for u.index < len(u.underlay) {
		flow = u.underlay[u.index].user.NextExepcted()
		if flow != rt.LoginFlow_None {
			return
		}

		u.index++
	}

	return
}
func (u *User) SetPassword(string) {}
func (u *User) SetTOTPCode(string) {}
func (u *User) SetToken(string)    {}
func (u *User) SetUsername(string) {}

type singleSpec struct {
	rs.BaseField

	Config map[string]publisher.Config `yaml:",inline"`
}

func (spec *singleSpec) resolve() (_ publisher.Interface, _ publisher.User, err error) {
	if len(spec.Config) != 1 {
		err = fmt.Errorf("unexpected count of config items %d (want exact one config)", len(spec.Config))
		return
	}

	for _, cfg := range spec.Config {
		return cfg.Create()
	}

	panic("unreachable")
}
