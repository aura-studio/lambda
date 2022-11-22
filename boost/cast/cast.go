package cast

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// ToInt casts an interface to an int type.
func ToInt(i any) int {
	v, _ := ToIntE(i)
	return v
}

// ToIntE casts an interface to an int type.
func ToIntE(a any) (int, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
		}
		return int(v.Int64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
		}
		n, _ := v.Int64()
		return int(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
		}
		n, _ := v.Float64()
		return int(n), nil
	case complex64:
		return int(real(v)), nil
	case complex128:
		return int(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		return int(v.UnixNano()), nil
	case time.Duration:
		return int(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			return int(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			return int(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			return int(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			return int(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	}
}

// ToInt8 casts an interface to an int8 type.
func ToInt8(i any) int8 {
	v, _ := ToInt8E(i)
	return v
}

// ToInt8E casts an interface to an int8 type.
func ToInt8E(a any) (int8, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return int8(v), nil
	case int8:
		return v, nil
	case int16:
		return int8(v), nil
	case int32:
		return int8(v), nil
	case int64:
		return int8(v), nil
	case uint:
		return int8(v), nil
	case uint8:
		return int8(v), nil
	case uint16:
		return int8(v), nil
	case uint32:
		return int8(v), nil
	case uint64:
		return int8(v), nil
	case float32:
		return int8(v), nil
	case float64:
		return int8(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
		}
		return int8(v.Int64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
		}
		n, _ := v.Int64()
		return int8(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
		}
		n, _ := v.Float64()
		return int8(n), nil
	case complex64:
		return int8(real(v)), nil
	case complex128:
		return int8(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		return int8(v.UnixNano()), nil
	case time.Duration:
		return int8(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			return int8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			return int8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			return int8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			return int8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	}
}

// ToInt16 casts an interface to an int16 type.
func ToInt16(i any) int16 {
	v, _ := ToInt16E(i)
	return v
}

// ToInt16E casts an interface to an int16 type.
func ToInt16E(a any) (int16, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return int16(v), nil
	case int8:
		return int16(v), nil
	case int16:
		return v, nil
	case int32:
		return int16(v), nil
	case int64:
		return int16(v), nil
	case uint:
		return int16(v), nil
	case uint8:
		return int16(v), nil
	case uint16:
		return int16(v), nil
	case uint32:
		return int16(v), nil
	case uint64:
		return int16(v), nil
	case float32:
		return int16(v), nil
	case float64:
		return int16(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
		}
		return int16(v.Int64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
		}
		n, _ := v.Int64()
		return int16(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
		}
		n, _ := v.Float64()
		return int16(n), nil
	case complex64:
		return int16(real(v)), nil
	case complex128:
		return int16(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		return int16(v.UnixNano()), nil
	case time.Duration:
		return int16(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			return int16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			return int16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			return int16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			return int16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	}
}

// ToInt32 casts an interface to an int32 type.
func ToInt32(i any) int32 {
	v, _ := ToInt32E(i)
	return v
}

// ToInt32E casts an interface to an int32 type.
func ToInt32E(a any) (int32, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return int32(v), nil
	case int8:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case uint:
		return int32(v), nil
	case uint8:
		return int32(v), nil
	case uint16:
		return int32(v), nil
	case uint32:
		return int32(v), nil
	case uint64:
		return int32(v), nil
	case float32:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
		}
		return int32(v.Int64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
		}
		n, _ := v.Int64()
		return int32(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
		}
		n, _ := v.Float64()
		return int32(n), nil
	case complex64:
		return int32(real(v)), nil
	case complex128:
		return int32(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		return int32(v.UnixNano()), nil
	case time.Duration:
		return int32(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			return int32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			return int32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			return int32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			return int32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	}
}

// ToInt64 casts an interface to an int64 type.
func ToInt64(i any) int64 {
	v, _ := ToInt64E(i)
	return v
}

// ToInt64E casts an interface to an int64 type.
func ToInt64E(a any) (int64, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
		}
		return v.Int64(), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
		}
		n, _ := v.Int64()
		return n, nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
		}
		n, _ := v.Float64()
		return int64(n), nil
	case complex64:
		return int64(real(v)), nil
	case complex128:
		return int64(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		return v.UnixNano(), nil
	case time.Duration:
		return int64(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	}
}

// ToUint casts an interface to a uint type.
func ToUint(i any) uint {
	v, _ := ToUintE(i)
	return v
}

// ToUintE casts an interface to a uint type.
func ToUintE(a any) (uint, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v), nil
	case uint:
		return v, nil
	case uint8:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v.Uint64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Uint64()
		return uint(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Float64()
		return uint(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		if v.UnixNano() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v.UnixNano()), nil
	case time.Duration:
		if int64(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	}
}

// ToUint8 casts an interface to a uint8 type.
func ToUint8(i any) uint8 {
	v, _ := ToUint8E(i)
	return v
}

// ToUint8E casts an interface to a uint type.
func ToUint8E(a any) (uint8, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v), nil
	case uint:
		return uint8(v), nil
	case uint8:
		return v, nil
	case uint16:
		return uint8(v), nil
	case uint32:
		return uint8(v), nil
	case uint64:
		return uint8(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v.Uint64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Uint64()
		return uint8(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Float64()
		return uint8(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		if v.UnixNano() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v.UnixNano()), nil
	case time.Duration:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint8(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	}
}

// ToUint16 casts an interface to a uint16 type.
func ToUint16(i any) uint16 {
	v, _ := ToUint16E(i)
	return v
}

// ToUint16E casts an interface to a uint16 type.
func ToUint16E(a any) (uint16, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v), nil
	case uint:
		return uint16(v), nil
	case uint8:
		return uint16(v), nil
	case uint16:
		return v, nil
	case uint32:
		return uint16(v), nil
	case uint64:
		return uint16(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v.Uint64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Uint64()
		return uint16(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Float64()
		return uint16(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		if v.UnixNano() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v.UnixNano()), nil
	case time.Duration:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint16(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	}
}

// ToUint32 casts an interface to a uint32 type.
func ToUint32(i any) uint32 {
	v, _ := ToUint32E(i)
	return v
}

// ToUint32E casts an interface to a uint32 type.
func ToUint32E(a any) (uint32, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v), nil
	case uint:
		return uint32(v), nil
	case uint8:
		return uint32(v), nil
	case uint16:
		return uint32(v), nil
	case uint32:
		return v, nil
	case uint64:
		return uint32(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v.Uint64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Uint64()
		return uint32(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Float64()
		return uint32(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		if v.UnixNano() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v.UnixNano()), nil
	case time.Duration:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint32(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	}
}

// ToUint64 casts an interface to a uint64 type.
func ToUint64(i any) uint64 {
	v, _ := ToUint64E(i)
	return v
}

// ToUint64E casts an interface to a uint64 type.
func ToUint64E(a any) (uint64, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return v.Uint64(), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Uint64()
		return n, nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		n, _ := v.Float64()
		return uint64(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		if v.UnixNano() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v.UnixNano()), nil
	case time.Duration:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
		}
		return uint64(v), nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint64(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint64(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint64(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			if n < 0 {
				return 0, fmt.Errorf("unable to cast %#v of type %T to negative value", v, v)
			}
			return uint64(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	}
}

// ToFloat32 casts an interface to a float32 type.
func ToFloat32(i any) float32 {
	v, _ := ToFloat32E(i)
	return v
}

// ToFloat32E casts an interface to a float32 type.
func ToFloat32E(a any) (float32, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return float32(v), nil
	case int8:
		return float32(v), nil
	case int16:
		return float32(v), nil
	case int32:
		return float32(v), nil
	case int64:
		return float32(v), nil
	case uint:
		return float32(v), nil
	case uint8:
		return float32(v), nil
	case uint16:
		return float32(v), nil
	case uint32:
		return float32(v), nil
	case uint64:
		return float32(v), nil
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
		}
		n, _ := new(big.Float).SetInt(v).Float32()
		return n, nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
		}
		n, _ := v.Float32()
		return n, nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
		}
		n, _ := v.Float32()
		return n, nil
	case complex64:
		return float32(real(v)), nil
	case complex128:
		return float32(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		return float32(v.UnixNano()), nil
	case time.Duration:
		return float32(v), nil
	case string:
		n, err := strconv.ParseFloat(v, 32)
		if err == nil {
			return float32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	case []byte:
		n, err := strconv.ParseFloat(string(v), 32)
		if err == nil {
			return float32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	case fmt.Stringer:
		n, err := strconv.ParseFloat(v.String(), 32)
		if err == nil {
			return float32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	case error:
		n, err := strconv.ParseFloat(v.Error(), 32)
		if err == nil {
			return float32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	}
}

// ToFloat64 casts an interface to a float64 type.
func ToFloat64(i any) float64 {
	v, _ := ToFloat64E(i)
	return v
}

// ToFloat64E casts an interface to a float64 type.
func ToFloat64E(a any) (float64, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
		}
		n, _ := new(big.Float).SetInt(v).Float64()
		return n, nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
		}
		n, _ := v.Float64()
		return n, nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
		}
		n, _ := v.Float64()
		return n, nil
	case complex64:
		return float64(real(v)), nil
	case complex128:
		return float64(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Time:
		return float64(v.UnixNano()), nil
	case time.Duration:
		return float64(v), nil
	case string:
		n, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	case []byte:
		n, err := strconv.ParseFloat(string(v), 64)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	case fmt.Stringer:
		n, err := strconv.ParseFloat(v.String(), 64)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	case error:
		n, err := strconv.ParseFloat(v.Error(), 64)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	}
}

// ToBigInt casts an interface to a *big.Int type.
func ToBigInt(i any) *big.Int {
	v, _ := ToBigIntE(i)
	return v
}

// ToBigIntE casts an interface to a *big.Int type.
func ToBigIntE(a any) (*big.Int, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int8:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int16:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int32:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int64:
		return big.NewInt(0).SetInt64(v), nil
	case uint:
		return big.NewInt(0).SetUint64(uint64(v)), nil
	case uint8:
		return big.NewInt(0).SetUint64(uint64(v)), nil
	case uint16:
		return big.NewInt(0).SetUint64(uint64(v)), nil
	case uint32:
		return big.NewInt(0).SetUint64(uint64(v)), nil
	case uint64:
		return big.NewInt(0).SetUint64(v), nil
	case float32:
		if math.IsInf(float64(v), 0) || math.IsNaN(float64(v)) {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to big.Float", a, a)
		}
		n, _ := big.NewFloat(float64(v)).Int(nil)
		return n, nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to big.Float", a, a)
		}
		n, _ := big.NewFloat(v).Int(nil)
		return n, nil
	case *big.Int:
		if v == nil {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
		}
		return v, nil
	case *big.Float:
		if v == nil {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
		}
		n, _ := v.Int(nil)
		return n, nil
	case *big.Rat:
		if v == nil {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
		}
		n, _ := v.Float64()
		return big.NewInt(int64(n)), nil
	case complex64:
		return big.NewInt(0).SetInt64(int64(real(v))), nil
	case complex128:
		return big.NewInt(0).SetInt64(int64(real(v))), nil
	case bool:
		if v {
			return big.NewInt(1), nil
		}
		return big.NewInt(0), nil
	case time.Time:
		return big.NewInt(0).SetInt64(v.UnixNano()), nil
	case time.Duration:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case string:
		n, ok := new(big.Int).SetString(v, 0)
		if ok {
			return n, nil
		}
		return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
	case []byte:
		n, ok := new(big.Int).SetString(string(v), 0)
		if ok {
			return n, nil
		}
		return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
	case fmt.Stringer:
		n, ok := new(big.Int).SetString(v.String(), 0)
		if ok {
			return n, nil
		}
		return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
	case error:
		n, ok := new(big.Int).SetString(v.Error(), 0)
		if ok {
			return n, nil
		}
		return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
	case nil:
		return big.NewInt(0), nil
	default:
		return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
	}
}

// ToBigFloat casts an interface to a *big.Float type.
func ToBigFloat(i any) *big.Float {
	v, _ := ToBigFloatE(i)
	return v
}

// ToBigFloatE casts an interface to a *big.Float type.
func ToBigFloatE(a any) (*big.Float, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case int8:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case int16:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case int32:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case int64:
		return big.NewFloat(0).SetInt64(v), nil
	case uint:
		return big.NewFloat(0).SetUint64(uint64(v)), nil
	case uint8:
		return big.NewFloat(0).SetUint64(uint64(v)), nil
	case uint16:
		return big.NewFloat(0).SetUint64(uint64(v)), nil
	case uint32:
		return big.NewFloat(0).SetUint64(uint64(v)), nil
	case uint64:
		return big.NewFloat(0).SetUint64(v), nil
	case float32:
		if math.IsInf(float64(v), 0) || math.IsNaN(float64(v)) {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to big.Float", a, a)
		}
		return big.NewFloat(float64(v)), nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to big.Float", a, a)
		}
		return big.NewFloat(v), nil
	case *big.Int:
		if v == nil {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
		}
		return big.NewFloat(0).SetInt(v), nil
	case *big.Float:
		if v == nil {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
		}
		return v, nil
	case *big.Rat:
		if v == nil {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
		}
		n, _ := v.Float64()
		return big.NewFloat(n), nil
	case complex64:
		return big.NewFloat(0).SetInt64(int64(real(v))), nil
	case complex128:
		return big.NewFloat(0).SetInt64(int64(real(v))), nil
	case bool:
		if v {
			return big.NewFloat(1), nil
		}
		return big.NewFloat(0), nil
	case time.Time:
		return big.NewFloat(0).SetInt64(v.UnixNano()), nil
	case time.Duration:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case string:
		n, ok := new(big.Float).SetString(v)
		if ok {
			return n, nil
		}
		return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
	case []byte:
		n, ok := new(big.Float).SetString(string(v))
		if ok {
			return n, nil
		}
		return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
	case fmt.Stringer:
		n, ok := new(big.Float).SetString(v.String())
		if ok {
			return n, nil
		}
		return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
	case error:
		n, ok := new(big.Float).SetString(v.Error())
		if ok {
			return n, nil
		}
		return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
	case nil:
		return big.NewFloat(0), nil
	default:
		return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
	}
}

// ToBigRat casts an interface to a *big.Rat type.
func ToBigRat(i any) *big.Rat {
	v, _ := ToBigRatE(i)
	return v
}

// ToBigRatE casts an interface to a *big.Rat type.
func ToBigRatE(a any) (*big.Rat, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return big.NewRat(int64(v), 1), nil
	case int8:
		return big.NewRat(int64(v), 1), nil
	case int16:
		return big.NewRat(int64(v), 1), nil
	case int32:
		return big.NewRat(int64(v), 1), nil
	case int64:
		return big.NewRat(v, 1), nil
	case uint:
		return big.NewRat(int64(v), 1), nil
	case uint8:
		return big.NewRat(int64(v), 1), nil
	case uint16:
		return big.NewRat(int64(v), 1), nil
	case uint32:
		return big.NewRat(int64(v), 1), nil
	case uint64:
		return big.NewRat(int64(v), 1), nil
	case float32:
		if math.IsInf(float64(v), 0) || math.IsNaN(float64(v)) {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to big.Rat", a, a)
		}
		return big.NewRat(0, 1).SetFloat64(float64(v)), nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to big.Rat", a, a)
		}
		return big.NewRat(0, 1).SetFloat64(v), nil
	case *big.Int:
		if v == nil {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
		}
		return big.NewRat(0, 1).SetFloat64(float64(v.Int64())), nil
	case *big.Float:
		if v == nil {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
		}
		n, _ := v.Float64()
		return big.NewRat(0, 1).SetFloat64(n), nil
	case *big.Rat:
		if v == nil {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
		}
		return v, nil
	case complex64:
		return big.NewRat(0, 1).SetFloat64(float64(real(v))), nil
	case complex128:
		return big.NewRat(0, 1).SetFloat64(real(v)), nil
	case bool:
		if v {
			return big.NewRat(1, 1), nil
		}
		return big.NewRat(0, 1), nil
	case time.Time:
		return big.NewRat(0, 1).SetInt64(v.UnixNano()), nil
	case time.Duration:
		return big.NewRat(0, 1).SetInt64(int64(v)), nil
	case string:
		n, ok := new(big.Rat).SetString(v)
		if ok {
			return n, nil
		}
		return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
	case []byte:
		n, ok := new(big.Rat).SetString(string(v))
		if ok {
			return n, nil
		}
		return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
	case fmt.Stringer:
		n, ok := new(big.Rat).SetString(v.String())
		if ok {
			return n, nil
		}
		return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
	case error:
		n, ok := new(big.Rat).SetString(v.Error())
		if ok {
			return n, nil
		}
		return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
	case nil:
		return big.NewRat(0, 1), nil
	default:
		return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
	}
}

// ToComplex64 casts an interface to a complex64 type.
func ToComplex64(i any) complex64 {
	v, _ := ToComplex64E(i)
	return v
}

// ToComplex64E casts an interface to a complex64 type.
func ToComplex64E(a any) (complex64, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return complex(float32(v), 0), nil
	case int8:
		return complex(float32(v), 0), nil
	case int16:
		return complex(float32(v), 0), nil
	case int32:
		return complex(float32(v), 0), nil
	case int64:
		return complex(float32(v), 0), nil
	case uint:
		return complex(float32(v), 0), nil
	case uint8:
		return complex(float32(v), 0), nil
	case uint16:
		return complex(float32(v), 0), nil
	case uint32:
		return complex(float32(v), 0), nil
	case uint64:
		return complex(float32(v), 0), nil
	case float32:
		return complex(v, 0), nil
	case float64:
		return complex(float32(v), 0), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
		}
		n, _ := new(big.Float).SetInt(v).Float32()
		return complex(n, 0), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
		}
		n, _ := v.Float32()
		return complex(n, 0), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
		}
		n, _ := new(big.Float).SetRat(v).Float32()
		return complex(n, 0), nil
	case complex64:
		return v, nil
	case complex128:
		return complex64(v), nil
	case bool:
		if v {
			return complex(1, 0), nil
		}
		return complex(0, 0), nil
	case time.Time:
		return complex(float32(v.UnixNano()), 0), nil
	case time.Duration:
		return complex(float32(v), 0), nil
	case string:
		n, err := strconv.ParseFloat(v, 32)
		if err == nil {
			return complex(float32(n), 0), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
	case []byte:
		n, err := strconv.ParseFloat(string(v), 32)
		if err == nil {
			return complex(float32(n), 0), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
	case fmt.Stringer:
		n, err := strconv.ParseFloat(v.String(), 32)
		if err == nil {
			return complex(float32(n), 0), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
	case error:
		n, err := strconv.ParseFloat(v.Error(), 32)
		if err == nil {
			return complex(float32(n), 0), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
	case nil:
		return complex(0, 0), nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
	}
}

// ToComplex128 casts an interface to a complex128 type.
func ToComplex128(i any) complex128 {
	v, _ := ToComplex128E(i)
	return v
}

// ToComplex128E casts an interface to a complex128 type.
func ToComplex128E(a any) (complex128, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return complex(float64(v), 0), nil
	case int8:
		return complex(float64(v), 0), nil
	case int16:
		return complex(float64(v), 0), nil
	case int32:
		return complex(float64(v), 0), nil
	case int64:
		return complex(float64(v), 0), nil
	case uint:
		return complex(float64(v), 0), nil
	case uint8:
		return complex(float64(v), 0), nil
	case uint16:
		return complex(float64(v), 0), nil
	case uint32:
		return complex(float64(v), 0), nil
	case uint64:
		return complex(float64(v), 0), nil
	case float32:
		f := float64(v)
		c := complex(f, 0)
		return c, nil
	case float64:
		return complex(v, 0), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
		}
		n, _ := new(big.Float).SetInt(v).Float64()
		return complex(n, 0), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
		}
		n, _ := v.Float64()
		return complex(n, 0), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
		}
		n, _ := new(big.Float).SetRat(v).Float64()
		return complex(n, 0), nil
	case complex64:
		return complex128(v), nil
	case complex128:
		return v, nil
	case bool:
		if v {
			return complex(1, 0), nil
		}
		return complex(0, 0), nil
	case time.Time:
		return complex(float64(v.UnixNano()), 0), nil
	case time.Duration:
		return complex(float64(v), 0), nil
	case string:
		n, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return complex(n, 0), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
	case []byte:
		n, err := strconv.ParseFloat(string(v), 64)
		if err == nil {
			return complex(n, 0), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
	case fmt.Stringer:
		n, err := strconv.ParseFloat(v.String(), 64)
		if err == nil {
			return complex(n, 0), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
	case error:
		n, err := strconv.ParseFloat(v.Error(), 64)
		if err == nil {
			return complex(n, 0), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
	case nil:
		return complex(0, 0), nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
	}
}

// ToBool casts an interface to a bool type.
func ToBool(i any) bool {
	v, _ := ToBoolE(i)
	return v
}

// ToBoolE casts an interface to a bool type.
func ToBoolE(a any) (bool, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case float32:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case *big.Int:
		if v == nil {
			return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
		}
		return v.Sign() != 0, nil
	case *big.Float:
		if v == nil {
			return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
		}
		return v.Sign() != 0, nil
	case *big.Rat:
		if v == nil {
			return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
		}
		return v.Sign() != 0, nil
	case complex64:
		return real(v) != 0 || imag(v) != 0, nil
	case complex128:
		return real(v) != 0 || imag(v) != 0, nil
	case bool:
		return v, nil
	case time.Time:
		return !v.IsZero(), nil
	case time.Duration:
		return v != 0, nil
	case string:
		n, err := parseInt(v)
		if err == nil {
			return n != 0, nil
		}
		t, err := strconv.ParseBool(v)
		if err == nil {
			return t, nil
		}
		return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
	case []byte:
		n, err := parseInt(string(v))
		if err == nil {
			return n != 0, nil
		}
		t, err := strconv.ParseBool(string(v))
		if err == nil {
			return t, nil
		}
		return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
	case fmt.Stringer:
		n, err := parseInt(v.String())
		if err == nil {
			return n != 0, nil
		}
		t, err := strconv.ParseBool(v.String())
		if err == nil {
			return t, nil
		}
		return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
	case error:
		n, err := parseInt(v.Error())
		if err == nil {
			return n != 0, nil
		}
		t, err := strconv.ParseBool(v.Error())
		if err == nil {
			return t, nil
		}
		return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
	case nil:
		return false, nil
	default:
		return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
	}
}

// ToString casts an interface to a string type.
func ToString(a any) string {
	v, _ := ToStringE(a)
	return v
}

// ToStringE casts an interface to a string type.
func ToStringE(a any) (string, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return strconv.Itoa(v), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.Itoa(int(v)), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(v), 10), nil
	case float32:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		//		return strconv.FormatFloat(float64(s), 'f', -1, 32), nil
		return decimal.NewFromFloat32(v).String(), nil
	case float64:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		// 		return strconv.FormatFloat(s, 'f', -1, 64), nil
		return decimal.NewFromFloat(v).String(), nil
	case *big.Int:
		return v.String(), nil
	case *big.Float:
		return v.String(), nil
	case *big.Rat:
		return v.String(), nil
	case complex64:
		return fmt.Sprintf("(%v+%vi)", real(v), imag(v)), nil
	case complex128:
		return fmt.Sprintf("(%v+%vi)", real(v), imag(v)), nil
	case bool:
		return strconv.FormatBool(v), nil
	case time.Time:
		return v.String(), nil
	case time.Duration:
		return v.String(), nil
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	case error:
		return v.Error(), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("unable to cast %#v of type %T to string", a, a)
	}
}

// ToBytes casts an interface to a []byte type.
func ToBytes(a any) []byte {
	v, _ := ToBytesE(a)
	return v
}

// ToBytesE casts an interface to a []byte type.
func ToBytesE(a any) ([]byte, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return []byte(strconv.Itoa(v)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return []byte(strconv.Itoa(int(v))), nil
	case int64:
		return []byte(strconv.FormatInt(v, 10)), nil
	case uint:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case float32:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		//		return []byte(strconv.FormatFloat(float64(s), 'f', -1, 32)), nil
		return []byte(decimal.NewFromFloat32(v).String()), nil
	case float64:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		// 		return []byte(strconv.FormatFloat(s, 'f', -1, 64)), nil
		return []byte(decimal.NewFromFloat(v).String()), nil
	case *big.Int:
		return []byte(v.String()), nil
	case *big.Float:
		return []byte(v.String()), nil
	case *big.Rat:
		return []byte(v.String()), nil
	case complex64:
		return []byte(fmt.Sprintf("(%v+%vi)", real(v), imag(v))), nil
	case complex128:
		return []byte(fmt.Sprintf("(%v+%vi)", real(v), imag(v))), nil
	case bool:
		return []byte(strconv.FormatBool(v)), nil
	case time.Time:
		return []byte(v.String()), nil
	case time.Duration:
		return []byte(v.String()), nil
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case fmt.Stringer:
		return []byte(v.String()), nil
	case error:
		return []byte(v.Error()), nil
	case nil:
		return []byte{}, nil
	default:
		return []byte{}, fmt.Errorf("unable to cast %#v of type %T to []byte", a, a)
	}
}

// ToStringer casts an interface to a fmt.Stringer type.
func ToStringer(a any) fmt.Stringer {
	v, _ := ToStringerE(a)
	return v
}

// ToStringerE casts an interface to a fmt.Stringer type.
func ToStringerE(a any) (fmt.Stringer, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return stringer{strconv.Itoa(v)}, nil
	case int8:
		return stringer{strconv.FormatInt(int64(v), 10)}, nil
	case int16:
		return stringer{strconv.FormatInt(int64(v), 10)}, nil
	case int32:
		return stringer{strconv.Itoa(int(v))}, nil
	case int64:
		return stringer{strconv.FormatInt(v, 10)}, nil
	case uint:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case uint8:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case uint16:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case uint32:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case uint64:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case float32:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		//		return stringer{strconv.FormatFloat(float64(s), 'f', -1, 32)}, nil
		return stringer{decimal.NewFromFloat32(v).String()}, nil
	case float64:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		// 		return stringer{strconv.FormatFloat(s, 'f', -1, 64)}, nil
		return stringer{decimal.NewFromFloat(v).String()}, nil
	case *big.Int:
		return stringer{v.String()}, nil
	case *big.Float:
		return stringer{v.String()}, nil
	case *big.Rat:
		return stringer{v.String()}, nil
	case complex64:
		return stringer{fmt.Sprintf("(%v+%vi)", real(v), imag(v))}, nil
	case complex128:
		return stringer{fmt.Sprintf("(%v+%vi)", real(v), imag(v))}, nil
	case bool:
		return stringer{strconv.FormatBool(v)}, nil
	case time.Time:
		return stringer{v.String()}, nil
	case time.Duration:
		return stringer{v.String()}, nil
	case string:
		return stringer{v}, nil
	case []byte:
		return stringer{string(v)}, nil
	case fmt.Stringer:
		return v, nil
	case error:
		return stringer{v.Error()}, nil
	case nil:
		return stringer{""}, nil
	default:
		return nil, fmt.Errorf("unable to cast %#v of type %T to fmt.Stringer", a, a)
	}
}

// ToError casts an interface to an error type.
func ToError(a any) error {
	v, _ := ToErrorE(a)
	return v
}

// ToErrorE casts an interface to an error type.
func ToErrorE(a any) (error, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return errors.New(strconv.Itoa(v)), nil
	case int8:
		return errors.New(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return errors.New(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return errors.New(strconv.Itoa(int(v))), nil
	case int64:
		return errors.New(strconv.FormatInt(v, 10)), nil
	case uint:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case float32:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		//		return errors.New(strconv.FormatFloat(float64(s), 'f', -1, 32)), nil
		return errors.New(decimal.NewFromFloat32(v).String()), nil
	case float64:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		// 		return errors.New(strconv.FormatFloat(s, 'f', -1, 64)), nil
		return errors.New(decimal.NewFromFloat(v).String()), nil
	case *big.Int:
		return errors.New(v.String()), nil
	case *big.Float:
		return errors.New(v.String()), nil
	case *big.Rat:
		return errors.New(v.String()), nil
	case complex64:
		return errors.New(fmt.Sprintf("(%v+%vi)", real(v), imag(v))), nil
	case complex128:
		return errors.New(fmt.Sprintf("(%v+%vi)", real(v), imag(v))), nil
	case bool:
		return errors.New(strconv.FormatBool(v)), nil
	case time.Time:
		return errors.New(v.String()), nil
	case time.Duration:
		return errors.New(v.String()), nil
	case string:
		return errors.New(v), nil
	case []byte:
		return errors.New(string(v)), nil
	case fmt.Stringer:
		return errors.New(v.String()), nil
	case error:
		return v, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unable to cast %#v of type %T to error", a, a)
	}
}

// ToTime casts an interface to a time.Time type.
func ToTime(a any, args ...any) time.Time {
	t, _ := ToTimeE(a)
	return t
}

// ToTimeE casts an interface to a time.Time type.
func ToTimeE(a any, args ...any) (time.Time, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return time.Unix(int64(v), 0), nil
	case int8:
		return time.Unix(int64(v), 0), nil
	case int16:
		return time.Unix(int64(v), 0), nil
	case int32:
		return time.Unix(int64(v), 0), nil
	case int64:
		return time.Unix(v, 0), nil
	case uint:
		return time.Unix(int64(v), 0), nil
	case uint8:
		return time.Unix(int64(v), 0), nil
	case uint16:
		return time.Unix(int64(v), 0), nil
	case uint32:
		return time.Unix(int64(v), 0), nil
	case uint64:
		return time.Unix(int64(v), 0), nil
	case float32:
		return time.Unix(int64(v), 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	case *big.Int:
		return time.Unix(v.Int64(), 0), nil
	case *big.Float:
		n, _ := v.Int64()
		return time.Unix(n, 0), nil
	case *big.Rat:
		n, _ := v.Float64()
		return time.Unix(int64(n), 0), nil
	case complex64:
		return time.Unix(int64(real(v)), 0), nil
	case complex128:
		return time.Unix(int64(real(v)), 0), nil
	case bool:
		if v {
			return time.Unix(1, 0), nil
		}
		return time.Unix(0, 0), nil
	case time.Time:
		return v, nil
	case time.Duration:
		return time.Unix(0, int64(v)), nil
	case string:
		return parseTime(v, args...)
	case []byte:
		return parseTime(string(v), args...)
	case fmt.Stringer:
		return parseTime(v.String(), args...)
	case nil:
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("unable to cast %#v of type %T to Time", a, a)
	}
}

// ToDuration casts an interface to a time.Duration type.
func ToDuration(i any, args ...any) time.Duration {
	d, _ := ToDurationE(i)
	return d
}

// ToDurationE casts an interface to a time.Duration type.
func ToDurationE(a any, args ...any) (d time.Duration, err error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return time.Duration(v), nil
	case int8:
		return time.Duration(v), nil
	case int16:
		return time.Duration(v), nil
	case int32:
		return time.Duration(v), nil
	case int64:
		return time.Duration(v), nil
	case uint:
		return time.Duration(v), nil
	case uint8:
		return time.Duration(v), nil
	case uint16:
		return time.Duration(v), nil
	case uint32:
		return time.Duration(v), nil
	case uint64:
		return time.Duration(v), nil
	case float32:
		return time.Duration(v), nil
	case float64:
		return time.Duration(v), nil
	case *big.Int:
		return time.Duration(v.Int64()), nil
	case *big.Float:
		n, _ := v.Int64()
		return time.Duration(n), nil
	case *big.Rat:
		n, _ := v.Float64()
		return time.Duration(n), nil
	case complex64:
		return time.Duration(real(v)), nil
	case complex128:
		return time.Duration(real(v)), nil
	case bool:
		if v {
			return time.Duration(1), nil
		}
		return time.Duration(0), nil
	case time.Time:
		return v.Sub(time.Time{}), nil
	case time.Duration:
		return v, nil
	case string:
		return parseDuration(v, args...)
	case []byte:
		return parseDuration(string(v), args...)
	case fmt.Stringer:
		return parseDuration(v.String(), args...)
	case error:
		return parseDuration(v.Error(), args...)
	default:
		return time.Duration(0), fmt.Errorf("unable to cast %#v of type %T to Duration", a, a)
	}
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a any) any {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirectToStringerOrError returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil) or an implementation of fmt.Stringer
// or error,
func indirectToStringerOrError(a any) any {
	if a == nil {
		return nil
	}

	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

type stringer struct{ string }

func (s stringer) String() string {
	return s.string
}

func parseInt(s string) (int64, error) {
	var foundZero bool
loop:
	for i := len(s); i > 0; i-- {
		switch s[i-1] {
		case '.':
			if foundZero {
				s = s[:i-1]
				break loop
			}
		case '0':
			foundZero = true
		default:
			break loop
		}
	}
	return strconv.ParseInt(s, 0, 0)
}

type TimeFormatType int

const (
	TimeFormatNoTimezone TimeFormatType = iota
	TimeFormatNamedTimezone
	TimeFormatNumericTimezone
	TimeFormatNumericAndNamedTimezone
	TimeFormatTimeOnly
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TimeFormatNoTimezone-0]
	_ = x[TimeFormatNamedTimezone-1]
	_ = x[TimeFormatNumericTimezone-2]
	_ = x[TimeFormatNumericAndNamedTimezone-3]
	_ = x[TimeFormatTimeOnly-4]
}

const _timeFormatType_name = "timeFormatNoTimezonetimeFormatNamedTimezonetimeFormatNumericTimezonetimeFormatNumericAndNamedTimezonetimeFormatTimeOnly"

var _timeFormatType_index = [...]uint8{0, 20, 43, 68, 101, 119}

func (i TimeFormatType) String() string {
	if i < 0 || i >= TimeFormatType(len(_timeFormatType_index)-1) {
		return "timeFormatType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _timeFormatType_name[_timeFormatType_index[i]:_timeFormatType_index[i+1]]
}

type TimeFormat struct {
	Format string
	Type   TimeFormatType
}

func (f TimeFormat) hasTimezone() bool {
	// We don't include the formats with only named timezones, see
	// https://github.com/golang/go/issues/19694#issuecomment-289103522
	return f.Type >= TimeFormatNumericTimezone && f.Type <= TimeFormatNumericAndNamedTimezone
}

var (
	defaultTimeFormats = []TimeFormat{
		{time.RFC3339, TimeFormatNumericTimezone},
		{"2006-01-02T15:04:05", TimeFormatNoTimezone}, // iso8601 without timezone
		{time.RFC1123Z, TimeFormatNumericTimezone},
		{time.RFC1123, TimeFormatNamedTimezone},
		{time.RFC822Z, TimeFormatNumericTimezone},
		{time.RFC822, TimeFormatNamedTimezone},
		{time.RFC850, TimeFormatNamedTimezone},
		{"2006-01-02 15:04:05.999999999 -0700 MST", TimeFormatNumericAndNamedTimezone}, // Time.String()
		{"2006-01-02T15:04:05-0700", TimeFormatNumericTimezone},                        // RFC3339 without timezone hh:mm colon
		{"2006-01-02 15:04:05Z0700", TimeFormatNumericTimezone},                        // RFC3339 without T or timezone hh:mm colon
		{"2006-01-02 15:04:05", TimeFormatNoTimezone},
		{time.ANSIC, TimeFormatNoTimezone},
		{time.UnixDate, TimeFormatNamedTimezone},
		{time.RubyDate, TimeFormatNumericTimezone},
		{"2006-01-02 15:04:05Z07:00", TimeFormatNumericTimezone},
		{"2006-01-02", TimeFormatNoTimezone},
		{"02 Jan 2006", TimeFormatNoTimezone},
		{"2006-01-02 15:04:05 -07:00", TimeFormatNumericTimezone},
		{"2006-01-02 15:04:05 -0700", TimeFormatNumericTimezone},
		{time.Kitchen, TimeFormatTimeOnly},
		{time.Stamp, TimeFormatTimeOnly},
		{time.StampMilli, TimeFormatTimeOnly},
		{time.StampMicro, TimeFormatTimeOnly},
		{time.StampNano, TimeFormatTimeOnly},
	}

	location = time.UTC
)

func parseTime(s string, args ...any) (d time.Time, e error) {
	var (
		location    *time.Location
		timeFormats []TimeFormat
	)
	for _, arg := range args {
		switch v := arg.(type) {
		case *time.Location:
			location = v
		case TimeFormat:
			timeFormats = append(timeFormats, v)
		}
	}

	if location == nil {
		location = time.Local
	}
	if len(timeFormats) == 0 {
		timeFormats = defaultTimeFormats
	}

	for _, timeFormat := range timeFormats {
		if d, e = time.Parse(timeFormat.Format, s); e == nil {

			// Some time formats have a zone name, but no offset, so it gets
			// put in that zone name (not the default one passed in to us), but
			// without that zone's offset. So set the location manually.
			if timeFormat.Type <= TimeFormatNamedTimezone {
				if location == nil {
					location = time.Local
				}
				year, month, day := d.Date()
				hour, min, sec := d.Clock()
				d = time.Date(year, month, day, hour, min, sec, d.Nanosecond(), location)
			}

			return
		}
	}
	return d, fmt.Errorf("unable to parse date: %s", s)
}

var (
	durationRegExp       *regexp.Regexp
	durationRegExpGroups = []string{
		`<years>[\+|\-]?\d+y`,
		`<months>[\+|\-]?\d+M`,
		`<days>[\+|\-]?\d+d`,
		`<hours>[\+|\-]?\d+h`,
		`<minutes>[\+|\-]?\d+m`,
		`<seconds>[\+|\-]?\d+s`,
		`<milliseconds>[\+|\-]?\d+ms`,
		`<microseconds>[\+|\-]?\d+us`,
		`<nanoseconds>[\+|\-]?\d+ns`,
	}
)

func init() {
	var buf = new(bytes.Buffer)
	for _, group := range durationRegExpGroups {
		buf.WriteString(`(?P`)
		buf.WriteString(group)
		buf.WriteString(`)?`)
	}
	durationRegExp = regexp.MustCompile(buf.String())
}

func parseDuration(s string, args ...any) (time.Duration, error) {
	var (
		t = time.Now()
	)

	for _, arg := range args {
		switch v := arg.(type) {
		case time.Time:
			t = v
		}
	}

	matches := durationRegExp.FindStringSubmatch(s)
	if len(matches) == 0 {
		return 0, nil
	}

	nums := []int{}
	for index := 1; index < len(matches); index++ {
		s := matches[index]
		if len(s) == 0 {
			nums = append(nums, 0)
			continue
		}
		for s[len(s)-1] < '0' || s[len(s)-1] > '9' {
			s = s[:len(s)-1]
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		nums = append(nums, int(n))
	}

	if len(nums) == 0 {
		return 0, fmt.Errorf("parse duration `%s` failed", s)
	}

	duration := t.AddDate(nums[0], nums[1], nums[2]).Add(
		time.Duration(nums[3]) * time.Hour,
	).Add(
		time.Duration(nums[4]) * time.Minute,
	).Add(
		time.Duration(nums[5]) * time.Second,
	).Add(
		time.Duration(nums[6]) * time.Millisecond,
	).Add(
		time.Duration(nums[7]) * time.Microsecond,
	).Add(
		time.Duration(nums[8]) * time.Nanosecond,
	).Sub(t)

	return duration, nil
}
