package encoding

import (
	"github.com/aura-studio/lambda/boost/ref"

	"gopkg.in/yaml.v2"
)

type YAML struct{}

func init() {
	register(new(YAML))
}

func NewYAML() *YAML {
	return new(YAML)
}

func (yaml YAML) String() string {
	return ref.TypeName(yaml)
}

func (yaml YAML) Style() EncodingStyleType {
	return EncodingStyleStruct
}

func (YAML) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (YAML) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func (yaml YAML) Reverse() Encoding {
	return yaml
}
