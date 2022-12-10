package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	"github.com/aura-studio/testdynamic1"
)

var tunnel = testdynamic1.Tunnel

//export dynamic_cgo_testdynamic1_test_init
func dynamic_cgo_testdynamic1_test_init() {
	tunnel.Init()
}

//export dynamic_cgo_testdynamic1_test_invoke
func dynamic_cgo_testdynamic1_test_invoke(route_cstr *C.char, req_cstr *C.char) *C.char {
	route := C.GoString(route_cstr)
	C.free(unsafe.Pointer(route_cstr))

	req := C.GoString(req_cstr)
	C.free(unsafe.Pointer(req_cstr))

	return C.CString(tunnel.Invoke(route, req))
}

//export dynamic_cgo_testdynamic1_test_close
func dynamic_cgo_testdynamic1_test_close() {
	tunnel.Close()
}

func main() {}
