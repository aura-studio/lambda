package encoding

import (
	"errors"
	"fmt"

	"github.com/aura-studio/lambda/boost/ref"
)

var (
	ErrEncodingMissingEncoding = errors.New("encoding cannot find encoding by name")
)

type EncodingStyleType int

const (
	EncodingStyleStruct = iota
	EncodingStyleBytes
	EncodingStyleMix
)

var encodingStyleName = map[EncodingStyleType]string{
	EncodingStyleStruct: "struct",
	EncodingStyleBytes:  "bytes",
	EncodingStyleMix:    "mix",
}

func (t EncodingStyleType) String() string {
	if s, ok := encodingStyleName[t]; ok {
		return s
	}
	return fmt.Sprintf("encodingStyleName=%d?", int(t))
}

type Encoding interface {
	String() string
	Style() EncodingStyleType
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
	Reverse() Encoding
}

func Empty() Encoding {
	return empty
}

type Bytes struct {
	Data []byte
}

var nilBytes = MakeBytes(nil)

func NewBytes() *Bytes {
	return new(Bytes)
}

func MakeBytes(v interface{}) Bytes {
	switch v := v.(type) {
	case []byte:
		return Bytes{
			Data: v,
		}
	case string:
		return Bytes{
			Data: []byte(v),
		}
	case Bytes:
		return v.Dulplicate()
	case *Bytes:
		return v.Dulplicate()
	default:
		return Bytes{
			Data: []byte(fmt.Sprint(v)),
		}
	}
}

func (b Bytes) Dulplicate() Bytes {
	data := make([]byte, len(b.Data))
	copy(data, b.Data)
	return MakeBytes(data)
}

func (b Bytes) Copy(in Bytes) {
	b.Data = make([]byte, len(in.Data))
	copy(b.Data, in.Data)
}

type EncodingSet map[string]Encoding

func newEncodingSet() EncodingSet {
	return make(map[string]Encoding)
}

var encodingSet = newEncodingSet()

func register(encodings ...Encoding) {
	encodingSet.register(encodings...)
}

func (es EncodingSet) register(encodings ...Encoding) {
	for _, encoding := range encodings {
		es[ref.TypeName(encoding)] = encoding
	}
}

func localEncoding(name string) (Encoding, error) {
	return encodingSet.locateEncoding(name)
}

func (es EncodingSet) locateEncoding(name string) (Encoding, error) {
	if e, ok := es[name]; ok {
		return e, nil
	}
	return nil, ErrEncodingMissingEncoding
}

func Marshal(e Encoding, v interface{}) ([]byte, error) {
	return e.Marshal(v)
}

func Unmarshal(e Encoding, data []byte, v interface{}) error {
	return e.Unmarshal(data, v)
}

func Encode(e Encoding, v interface{}) []byte {
	data, err := e.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func Decode(e Encoding, data []byte, v interface{}) {
	if err := e.Unmarshal(data, v); err != nil {
		panic(err)
	}
}
