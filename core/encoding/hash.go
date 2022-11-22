package encoding

import (
	"encoding/json"

	"github.com/aura-studio/lambda/boost/ref"
)

type Hash struct{}

func init() {
	register(NewHash())
}

func NewHash() *Hash {
	return new(Hash)
}

func (h Hash) String() string {
	return ref.TypeName(h)
}

func (h Hash) Style() EncodingStyleType {
	return EncodingStyleStruct
}

func (Hash) Marshal(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case HashMarshaller:
		hash := v.MarshalHash()
		return json.Marshal(hash)
	}
	return nil, nil
}

func (Hash) Unmarshal(data []byte, v interface{}) error {
	pairs := make([][]interface{}, 0)
	if err := json.Unmarshal(data, &pairs); err != nil {
		return err
	}
	switch v := v.(type) {
	case HashUnmarshaller:
		v.UnmarshalHash(pairs)
		return nil
	}
	return nil

}

func (h Hash) Reverse() Encoding {
	return h
}

type HashMarshaller interface {
	MarshalHash() [][]interface{}
}

type HashUnmarshaller interface {
	UnmarshalHash(pairs [][]interface{})
}
