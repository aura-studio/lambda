package main

/*
#cgo CFLAGS: -I/tmp/warehouse/testdynamic1_test/
#cgo LDFLAGS: -L/tmp/warehouse/testdynamic1_test/ -lcgo_testdynamic1_test
#include "/tmp/warehouse/testdynamic1_test/libcgo_testdynamic1_test.h"
#include "stdlib.h"
*/
import "C"
import "unsafe"

type tunnel struct{}

func (t tunnel) Init() {
	C.dynamic_cgo_testdynamic1_test_init()
}

func (t tunnel) Invoke(route string, req string) string {
	rsp_cstr := C.dynamic_cgo_testdynamic1_test_invoke(C.CString(route), C.CString(req))
	rsp := C.GoString(rsp_cstr)
	C.free(unsafe.Pointer(rsp_cstr))
	return rsp
}

func (t tunnel) Close() {
	C.dynamic_cgo_testdynamic1_test_close()
}

var Tunnel tunnel

func main() {}
