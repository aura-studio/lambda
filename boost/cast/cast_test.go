package cast_test

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/aura-studio/lambda/boost/cast"

	. "github.com/frankban/quicktest"
)

type testStep struct {
	input  any
	expect any
	iserr  bool
}

type testValues struct {
	zero, one, eight, eightnegative, eightpoint31, eightpoint31negative any
}

type stringer struct{ string }

func (s stringer) String() string {
	return s.string
}

func createTestSteps(t testValues) []testStep {
	kind := reflect.TypeOf(t.zero).Kind()
	isUint := kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64

	// Some precision is lost when converting from float64 to float32.
	eightpoint31_32 := t.eightpoint31
	eightpoint31negative_32 := t.eightpoint31negative
	if _, ok := t.zero.(float64); ok {
		eightpoint31_32 = float64(float32(t.eightpoint31.(float64)))
		eightpoint31negative_32 = float64(float32(t.eightpoint31negative.(float64)))
	}
	if _, ok := t.zero.(complex128); ok {
		eightpoint31_32 = complex128(complex64(t.eightpoint31.(complex128)))
		eightpoint31negative_32 = complex128(complex64(t.eightpoint31negative.(complex128)))
	}
	if _, ok := t.zero.(*big.Float); ok {
		f, _ := t.eightpoint31.(*big.Float).Float32()
		eightpoint31_32 = new(big.Float).SetFloat64(float64(f))
		f, _ = t.eightpoint31negative.(*big.Float).Float32()
		eightpoint31negative_32 = new(big.Float).SetFloat64(float64(f))
	}
	if _, ok := t.zero.(*big.Rat); ok {
		f, _ := t.eightpoint31.(*big.Rat).Float32()
		eightpoint31_32 = new(big.Rat).SetFloat64(float64(f))
		f, _ = t.eightpoint31negative.(*big.Rat).Float32()
		eightpoint31negative_32 = new(big.Rat).SetFloat64(float64(f))
	}

	return []testStep{
		{int(8), t.eight, false},
		{int8(8), t.eight, false},
		{int16(8), t.eight, false},
		{int32(8), t.eight, false},
		{int64(8), t.eight, false},
		{uint(8), t.eight, false},
		{uint8(8), t.eight, false},
		{uint16(8), t.eight, false},
		{uint32(8), t.eight, false},
		{uint64(8), t.eight, false},
		{float32(8.31), eightpoint31_32, false},
		{float64(8.31), t.eightpoint31, false},
		{big.NewInt(8), t.eight, false},
		{big.NewFloat(8.31), t.eightpoint31, false},
		{big.NewRat(8, 1), t.eight, false},
		{complex64(8 + 0i), t.eight, false},
		{complex128(8 + 0i), t.eight, false},
		{true, t.one, false},
		{false, t.zero, false},
		{"8", t.eight, false},
		{[]byte{56}, t.eight, false},
		{stringer{"8"}, t.eight, false},
		{errors.New("8"), t.eight, false},
		{time.Unix(0, 8), t.eight, false},
		{time.Duration(8), t.eight, false},
		{nil, t.zero, false},
		{int(-8), t.eightnegative, isUint},
		{int8(-8), t.eightnegative, isUint},
		{int16(-8), t.eightnegative, isUint},
		{int32(-8), t.eightnegative, isUint},
		{int64(-8), t.eightnegative, isUint},
		{float32(-8.31), eightpoint31negative_32, isUint},
		{float64(-8.31), t.eightpoint31negative, isUint},
		{big.NewInt(-8), t.eightnegative, isUint},
		{big.NewFloat(-8.31), t.eightpoint31negative, isUint},
		{big.NewRat(-8, 1), t.eightnegative, isUint},
		{complex64(-8 + 0i), t.eightnegative, isUint},
		{complex128(-8 + 0i), t.eightnegative, isUint},
		{"-8", t.eightnegative, isUint},
		{[]byte{45, 56}, t.eightnegative, isUint},
		{stringer{"-8"}, t.eightnegative, isUint},
		{errors.New("-8"), t.eightnegative, isUint},
		{time.Unix(0, -8), t.eightnegative, isUint},
		{time.Duration(-8), t.eightnegative, isUint},
		{"test", t.zero, true},
		{testing.T{}, t.zero, true},
	}
}

func runTest(c *C, tests []testStep, tove func(any) (any, error), tov func(any) any) {
	c.Helper()

	for i, test := range tests {
		errmsg := Commentf("i = %d, test = %#v", i, test)

		v, err := tove(test.input)
		if test.iserr {
			c.Assert(err, IsNotNil, errmsg)
			continue
		}
		c.Assert(err, IsNil, errmsg)
		t := reflect.TypeOf(v)
		if t.Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
			c.Assert(v.(fmt.Stringer).String(), Equals, test.expect.(fmt.Stringer).String(), errmsg)
		} else {
			c.Assert(v, Equals, test.expect, errmsg)
		}

		// Non-E test:
		v = tov(test.input)
		t = reflect.TypeOf(v)
		if t.Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
			c.Assert(v.(fmt.Stringer).String(), Equals, test.expect.(fmt.Stringer).String(), errmsg)
		} else {
			c.Assert(v, Equals, test.expect, errmsg)
		}
	}
}

func TestToIntE(t *testing.T) {
	tests := createTestSteps(testValues{
		int(0),
		int(1),
		int(8),
		int(-8),
		int(8),
		int(-8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToIntE(v) },
		func(v any) any { return cast.ToInt(v) },
	)
}

func TestToInt8E(t *testing.T) {
	tests := createTestSteps(testValues{
		int8(0),
		int8(1),
		int8(8),
		int8(-8),
		int8(8),
		int8(-8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToInt8E(v) },
		func(v any) any { return cast.ToInt8(v) },
	)
}

func TestToInt16E(t *testing.T) {
	tests := createTestSteps(testValues{
		int16(0),
		int16(1),
		int16(8),
		int16(-8),
		int16(8),
		int16(-8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToInt16E(v) },
		func(v any) any { return cast.ToInt16(v) },
	)
}

func TestToInt32E(t *testing.T) {
	tests := createTestSteps(testValues{
		int32(0),
		int32(1),
		int32(8),
		int32(-8),
		int32(8),
		int32(-8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToInt32E(v) },
		func(v any) any { return cast.ToInt32(v) },
	)
}

func TestToInt64E(t *testing.T) {
	tests := createTestSteps(testValues{
		int64(0),
		int64(1),
		int64(8),
		int64(-8),
		int64(8),
		int64(-8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToInt64E(v) },
		func(v any) any { return cast.ToInt64(v) },
	)
}

func TestToUintE(t *testing.T) {
	tests := createTestSteps(testValues{
		uint(0),
		uint(1),
		uint(8),
		uint(0),
		uint(8),
		uint(8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToUintE(v) },
		func(v any) any { return cast.ToUint(v) },
	)
}

func TestToUint8E(t *testing.T) {
	tests := createTestSteps(testValues{
		uint8(0),
		uint8(1),
		uint8(8),
		uint8(0),
		uint8(8),
		uint8(8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToUint8E(v) },
		func(v any) any { return cast.ToUint8(v) },
	)
}

func TestToUint16E(t *testing.T) {
	tests := createTestSteps(testValues{
		uint16(0),
		uint16(1),
		uint16(8),
		uint16(0),
		uint16(8),
		uint16(8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToUint16E(v) },
		func(v any) any { return cast.ToUint16(v) },
	)
}

func TestToUint32E(t *testing.T) {
	tests := createTestSteps(testValues{
		uint32(0),
		uint32(1),
		uint32(8),
		uint32(0),
		uint32(8),
		uint32(8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToUint32E(v) },
		func(v any) any { return cast.ToUint32(v) },
	)
}

func TestToUint64E(t *testing.T) {
	tests := createTestSteps(testValues{
		uint64(0),
		int64(1),
		uint64(8),
		uint64(0),
		uint64(8),
		uint64(8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToUint64E(v) },
		func(v any) any { return cast.ToUint64(v) },
	)
}

func TestToFloat32E(t *testing.T) {
	tests := createTestSteps(testValues{
		float32(0),
		float32(1),
		float32(8),
		float32(-8),
		float32(8.31),
		float32(-8.31),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToFloat32E(v) },
		func(v any) any { return cast.ToFloat32(v) },
	)
}

func TestToFloat64E(t *testing.T) {
	tests := createTestSteps(testValues{
		float64(0),
		float64(1),
		float64(8),
		float64(-8),
		float64(8.31),
		float64(-8.31),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToFloat64E(v) },
		func(v any) any { return cast.ToFloat64(v) },
	)
}

func TestToBigIntE(t *testing.T) {
	tests := createTestSteps(testValues{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(8),
		big.NewInt(-8),
		big.NewInt(8),
		big.NewInt(-8),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToBigIntE(v) },
		func(v any) any { return cast.ToBigInt(v) },
	)
}

func TestToBigFloatE(t *testing.T) {
	tests := createTestSteps(testValues{
		big.NewFloat(0),
		big.NewFloat(1),
		big.NewFloat(8),
		big.NewFloat(-8),
		big.NewFloat(8.31),
		big.NewFloat(-8.31),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToBigFloatE(v) },
		func(v any) any { return cast.ToBigFloat(v) },
	)
}

func TestToBigRatE(t *testing.T) {
	tests := createTestSteps(testValues{
		big.NewRat(0, 1),
		big.NewRat(1, 1),
		big.NewRat(8, 1),
		big.NewRat(-8, 1),
		big.NewRat(0, 1).SetFloat64(8.31),
		big.NewRat(0, 1).SetFloat64(-8.31),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToBigRatE(v) },
		func(v any) any { return cast.ToBigRat(v) },
	)
}

func TestToComplex64(t *testing.T) {
	tests := createTestSteps(testValues{
		complex64(0 + 0i),
		complex64(1 + 0i),
		complex64(8 + 0i),
		complex64(-8 + 0i),
		complex64(8.31 + 0i),
		complex64(-8.31 + 0i),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToComplex64E(v) },
		func(v any) any { return cast.ToComplex64(v) },
	)
}

func TestToComplex128(t *testing.T) {
	tests := createTestSteps(testValues{
		complex128(0 + 0i),
		complex128(1 + 0i),
		complex128(8 + 0i),
		complex128(-8 + 0i),
		complex128(8.31 + 0i),
		complex128(-8.31 + 0i),
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToComplex128E(v) },
		func(v any) any { return cast.ToComplex128(v) },
	)
}

func TestToBoolE(t *testing.T) {
	tests := createTestSteps(testValues{
		false,
		true,
		true,
		true,
		true,
		true,
	})

	runTest(
		New(t),
		tests,
		func(v any) (any, error) { return cast.ToBoolE(v) },
		func(v any) any { return cast.ToBool(v) },
	)
}

func TestToTimeE(t *testing.T) {
	c := New(t)

	tests := []struct {
		input  any
		expect time.Time
		iserr  bool
	}{
		{"2009-11-10 23:00:00 +0000 UTC", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},   // Time.String()
		{"Tue Nov 10 23:00:00 2009", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},        // ANSIC
		{"Tue Nov 10 23:00:00 UTC 2009", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},    // UnixDate
		{"Tue Nov 10 23:00:00 +0000 2009", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},  // RubyDate
		{"10 Nov 09 23:00 UTC", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},             // RFC822
		{"10 Nov 09 23:00 +0000", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},           // RFC822Z
		{"Tuesday, 10-Nov-09 23:00:00 UTC", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false}, // RFC850
		{"Tue, 10 Nov 2009 23:00:00 UTC", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},   // RFC1123
		{"Tue, 10 Nov 2009 23:00:00 +0000", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false}, // RFC1123Z
		{"2009-11-10T23:00:00Z", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},            // RFC3339
		{"2018-10-21T23:21:29+0200", time.Date(2018, 10, 21, 21, 21, 29, 0, time.UTC), false},      // RFC3339 without timezone hh:mm colon
		{"2009-11-10T23:00:00Z", time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), false},            // RFC3339Nano
		{"11:00PM", time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC), false},                              // Kitchen
		{"Nov 10 23:00:00", time.Date(0, 11, 10, 23, 0, 0, 0, time.UTC), false},                    // Stamp
		{"Nov 10 23:00:00.000", time.Date(0, 11, 10, 23, 0, 0, 0, time.UTC), false},                // StampMilli
		{"Nov 10 23:00:00.000000", time.Date(0, 11, 10, 23, 0, 0, 0, time.UTC), false},             // StampMicro
		{"Nov 10 23:00:00.000000000", time.Date(0, 11, 10, 23, 0, 0, 0, time.UTC), false},          // StampNano
		{"2016-03-06 15:28:01-00:00", time.Date(2016, 3, 6, 15, 28, 1, 0, time.UTC), false},        // RFC3339 without T
		{"2016-03-06 15:28:01-0000", time.Date(2016, 3, 6, 15, 28, 1, 0, time.UTC), false},         // RFC3339 without T or timezone hh:mm colon
		{"2016-03-06 15:28:01", time.Date(2016, 3, 6, 15, 28, 1, 0, time.UTC), false},
		{"2016-03-06 15:28:01 -0000", time.Date(2016, 3, 6, 15, 28, 1, 0, time.UTC), false},
		{"2016-03-06 15:28:01 -00:00", time.Date(2016, 3, 6, 15, 28, 1, 0, time.UTC), false},
		{"2016-03-06 15:28:01 +0900", time.Date(2016, 3, 6, 6, 28, 1, 0, time.UTC), false},
		{"2016-03-06 15:28:01 +09:00", time.Date(2016, 3, 6, 6, 28, 1, 0, time.UTC), false},
		{"2006-01-02", time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC), false},
		{"02 Jan 2006", time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC), false},
		{1472574600, time.Date(2016, 8, 30, 16, 30, 0, 0, time.UTC), false},
		{int(1482597504), time.Date(2016, 12, 24, 16, 38, 24, 0, time.UTC), false},
		{int64(1234567890), time.Date(2009, 2, 13, 23, 31, 30, 0, time.UTC), false},
		{int32(1234567890), time.Date(2009, 2, 13, 23, 31, 30, 0, time.UTC), false},
		{uint(1482597504), time.Date(2016, 12, 24, 16, 38, 24, 0, time.UTC), false},
		{uint64(1234567890), time.Date(2009, 2, 13, 23, 31, 30, 0, time.UTC), false},
		{uint32(1234567890), time.Date(2009, 2, 13, 23, 31, 30, 0, time.UTC), false},
		{time.Date(2009, 2, 13, 23, 31, 30, 0, time.UTC), time.Date(2009, 2, 13, 23, 31, 30, 0, time.UTC), false},
		// errors
		{"2006", time.Time{}, true},
		{testing.T{}, time.Time{}, true},
	}

	for i, test := range tests {
		errmsg := Commentf("i = %d", i) // assert helper message

		v, err := cast.ToTimeE(test.input)
		if test.iserr {
			c.Assert(err, IsNotNil)
			continue
		}

		c.Assert(err, IsNil)
		c.Assert(v.UTC(), Equals, test.expect, errmsg)

		// Non-E test
		v = cast.ToTime(test.input)
		c.Assert(v.UTC(), Equals, test.expect, errmsg)
	}
}

// func TestToStringE(t *testing.T) {
// 	c := New(t)

// 	var jn json.Number
// 	_ = json.Unmarshal([]byte("8"), &jn)
// 	type Key struct {
// 		k string
// 	}
// 	key := &Key{"foo"}

// 	tests := []struct {
// 		input  any
// 		expect string
// 		iserr  bool
// 	}{
// 		{int(8), "8", false},
// 		{int8(8), "8", false},
// 		{int16(8), "8", false},
// 		{int32(8), "8", false},
// 		{int64(8), "8", false},
// 		{uint(8), "8", false},
// 		{uint8(8), "8", false},
// 		{uint16(8), "8", false},
// 		{uint32(8), "8", false},
// 		{uint64(8), "8", false},
// 		{float32(8.31), "8.31", false},
// 		{float64(8.31), "8.31", false},
// 		{jn, "8", false},
// 		{true, "true", false},
// 		{false, "false", false},
// 		{nil, "", false},
// 		{[]byte("one time"), "one time", false},
// 		{"one more time", "one more time", false},
// 		{template.HTML("one time"), "one time", false},
// 		{template.URL("http://somehost.foo"), "http://somehost.foo", false},
// 		{template.JS("(1+2)"), "(1+2)", false},
// 		{template.CSS("a"), "a", false},
// 		{template.HTMLAttr("a"), "a", false},
// 		// errors
// 		{testing.T{}, "", true},
// 		{key, "", true},
// 	}

// 	for i, test := range tests {
// 		errmsg := Commentf("i = %d", i) // assert helper message

// 		v, err := ToStringE(test.input)
// 		if test.iserr {
// 			c.Assert(err, IsNotNil, errmsg)
// 			continue
// 		}

// 		c.Assert(err, IsNil, errmsg)
// 		c.Assert(v, Equals, test.expect, errmsg)

// 		// Non-E test
// 		v = ToString(test.input)
// 		c.Assert(v, Equals, test.expect, errmsg)
// 	}
// }

// type foo struct {
// 	val string
// }

// func (x foo) String() string {
// 	return x.val
// }

// func TestStringerToString(t *testing.T) {
// 	c := New(t)

// 	var x foo
// 	x.val = "bar"
// 	c.Assert(ToString(x), Equals, "bar")
// }

// type fu struct {
// 	val string
// }

// func (x fu) Error() string {
// 	return x.val
// }

// func TestErrorToString(t *testing.T) {
// 	c := New(t)

// 	var x fu
// 	x.val = "bar"
// 	c.Assert(ToString(x), Equals, "bar")
// }

// func TestToDurationE(t *testing.T) {
// 	c := New(t)

// 	var td time.Duration = 5
// 	var jn json.Number
// 	_ = json.Unmarshal([]byte("5"), &jn)

// 	tests := []struct {
// 		input  any
// 		expect time.Duration
// 		iserr  bool
// 	}{
// 		{time.Duration(5), td, false},
// 		{int(5), td, false},
// 		{int64(5), td, false},
// 		{int32(5), td, false},
// 		{int16(5), td, false},
// 		{int8(5), td, false},
// 		{uint(5), td, false},
// 		{uint64(5), td, false},
// 		{uint32(5), td, false},
// 		{uint16(5), td, false},
// 		{uint8(5), td, false},
// 		{float64(5), td, false},
// 		{float32(5), td, false},
// 		{jn, td, false},
// 		{string("5"), td, false},
// 		{string("5ns"), td, false},
// 		{string("5us"), time.Microsecond * td, false},
// 		{string("5µs"), time.Microsecond * td, false},
// 		{string("5ms"), time.Millisecond * td, false},
// 		{string("5s"), time.Second * td, false},
// 		{string("5m"), time.Minute * td, false},
// 		{string("5h"), time.Hour * td, false},
// 		// errors
// 		{"test", 0, true},
// 		{testing.T{}, 0, true},
// 	}

// 	for i, test := range tests {
// 		errmsg := Commentf("i = %d", i) // assert helper message

// 		v, err := ToDurationE(test.input)
// 		if test.iserr {
// 			c.Assert(err, IsNotNil)
// 			continue
// 		}

// 		c.Assert(err, IsNil)
// 		c.Assert(v, Equals, test.expect, errmsg)

// 		// Non-E test
// 		v = ToDuration(test.input)
// 		c.Assert(v, Equals, test.expect, errmsg)
// 	}
// }
