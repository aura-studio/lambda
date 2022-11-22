package ref_test

import (
	"fmt"
	"reflect"
	"testing"
)

type T struct {
	T2
}

func (t *T) Foo() {
	fmt.Println("foo")
}

type T2 struct{}

func (t *T2) Bar() {
	fmt.Println("bar")
}

func TestCallName(t *testing.T) {
	var o T
	reflect.ValueOf(&o).MethodByName("Foo").Call([]reflect.Value{})
	reflect.ValueOf(&o).MethodByName("Bar").Call([]reflect.Value{})
}
