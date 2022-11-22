package encoding

import (
	"encoding/json"
	"errors"

	"github.com/aura-studio/lambda/boost/ref"
)

var (
	ErrJSONWrongValueType = errors.New("encoding JSON converts on wrong type value")
)

type JSON struct{}

func init() {
	register(NewJSON())
}

func NewJSON() *JSON {
	return new(JSON)
}

func (json JSON) String() string {
	return ref.TypeName(json)
}

func (json JSON) Style() EncodingStyleType {
	return EncodingStyleStruct
}

func (JSON) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSON) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (json JSON) Reverse() Encoding {
	return json
}
