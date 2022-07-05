package rshelper

import (
	"reflect"

	"arhat.dev/rs"
)

func InitAll(f rs.Field, opts *rs.Options) rs.Field {
	rs.InitRecursively(reflect.ValueOf(f), opts)
	return f
}
