package encoding

import (
	"encoding/base64"
	"errors"

	"github.com/aura-studio/lambda/boost/ref"
)

var (
	ErrBase64WrongValueType    = errors.New("encoding base64 converts on wrong type value")
	ErrBase64URLWrongValueType = errors.New("encoding base64URL converts on wrong type value")
)

type Base64 struct{}

func init() {
	register(NewBase64())
}

func NewBase64() *Base64 {
	return new(Base64)
}

func (b Base64) String() string {
	return ref.TypeName(b)
}

func (Base64) Style() EncodingStyleType {
	return EncodingStyleBytes
}

func (Base64) Marshal(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case []byte:
		s := base64.StdEncoding.EncodeToString(v)
		return []byte(s), nil
	case Bytes:
		s := base64.StdEncoding.EncodeToString(v.Data)
		return []byte(s), nil
	case *Bytes:
		s := base64.StdEncoding.EncodeToString(v.Data)
		return []byte(s), nil
	default:
		return nil, ErrBase64WrongValueType
	}
}

func (Base64) Unmarshal(data []byte, v interface{}) error {
	switch v := v.(type) {
	case *Bytes:
		s, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return err
		}
		v.Data = []byte(s)
		return nil
	default:
		return ErrBase64WrongValueType
	}
}

func (b Base64) Reverse() Encoding {
	return b
}

type Base64URL struct{}

func init() {
	register(NewBase64URL())
}

func NewBase64URL() *Base64URL {
	return new(Base64URL)
}

func (b Base64URL) String() string {
	return ref.TypeName(b)
}

func (Base64URL) Style() EncodingStyleType {
	return EncodingStyleBytes
}

func (Base64URL) Marshal(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case []byte:
		s := base64.URLEncoding.EncodeToString(v)
		return []byte(s), nil
	case Bytes:
		s := base64.URLEncoding.EncodeToString(v.Data)
		return []byte(s), nil
	case *Bytes:
		s := base64.URLEncoding.EncodeToString(v.Data)
		return []byte(s), nil
	default:
		return nil, ErrBase64URLWrongValueType
	}
}

func (Base64URL) Unmarshal(data []byte, v interface{}) error {
	switch v := v.(type) {
	case *Bytes:
		s, err := base64.URLEncoding.DecodeString(string(data))
		if err != nil {
			return err
		}
		v.Data = []byte(s)
		return nil
	default:
		return ErrBase64URLWrongValueType
	}
}

func (b Base64URL) Reverse() Encoding {
	return b
}
