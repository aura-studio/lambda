package magic

import (
	"context"
	"reflect"
)

var (
	TypeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()
	TypeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	TypeOfBytes   = reflect.TypeOf(([]byte)(nil))
	TypeNil       = reflect.Type(nil)
)

