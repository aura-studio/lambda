package encoding

import (
	"errors"
	"strings"

	"github.com/aura-studio/lambda/boost/magic"
)

var (
	ErrWrongEncodingStyle = errors.New("wrong encoding style is found in chain")
)

type ChainEncoding struct {
	encoder []string
	decoder []string
}

func NewChainEncoding(encoder, decoder []string) *ChainEncoding {
	return &ChainEncoding{
		encoder: encoder,
		decoder: decoder,
	}
}

var empty *ChainEncoding = NewChainEncoding([]string{"Lazy"}, []string{"Lazy"})

func (c ChainEncoding) String() string {
	var builder strings.Builder
	builder.WriteString(magic.SeparatorBracketLeft)
	for index, name := range c.encoder {
		builder.WriteString(name)
		if index != len(c.encoder)-1 {
			builder.WriteString(magic.SeparatorColon)
		}
	}
	builder.WriteString(magic.SeparatorBracketRight)
	builder.WriteString(magic.SeparatorSpace)
	builder.WriteString(magic.SeparatorMinus)
	builder.WriteString(magic.SeparatorGreater)
	builder.WriteString(magic.SeparatorSpace)
	builder.WriteString(magic.SeparatorBracketLeft)
	for index, name := range c.decoder {
		builder.WriteString(name)
		if index != len(c.decoder)-1 {
			builder.WriteString(magic.SeparatorColon)
		}
	}
	builder.WriteString(magic.SeparatorBracketRight)
	return builder.String()
}

func (c ChainEncoding) Style() EncodingStyleType {
	return EncodingStyleMix
}

func (c ChainEncoding) Reverse() Encoding {
	re := ChainEncoding{
		encoder: make([]string, len(c.decoder)),
		decoder: make([]string, len(c.encoder)),
	}
	lenDecoder := len(c.decoder)
	for index := 0; index < lenDecoder; index++ {
		re.encoder[index] = c.decoder[lenDecoder-1-index]
	}
	lenEncoder := len(c.encoder)
	for index := 0; index < lenEncoder; index++ {
		re.decoder[index] = c.encoder[lenEncoder-1-index]
	}
	return re
}

func (c ChainEncoding) Marshal(v interface{}) ([]byte, error) {
	var data []byte
	for index, name := range c.encoder {
		encoding, err := localEncoding(name)
		if err != nil {
			return nil, err
		}
		if index == 0 {
			data, err = encoding.Marshal(v)
			if err != nil {
				return nil, err
			}
		} else {
			if encoding.Style() == EncodingStyleStruct {
				return nil, ErrWrongEncodingStyle
			}
			data, err = encoding.Marshal(data)
			if err != nil {
				return nil, err
			}
		}
	}
	return data, nil
}

func (c ChainEncoding) Unmarshal(data []byte, v interface{}) error {
	bytes := MakeBytes(nil)
	for index, name := range c.decoder {
		encoding, err := localEncoding(name)
		if err != nil {
			return err
		}
		if index < len(c.decoder)-1 {
			if encoding.Style() == EncodingStyleStruct {
				return ErrWrongEncodingStyle
			}
			err = encoding.Unmarshal(data, &bytes)
			if err != nil {
				return err
			}
			data = bytes.Data
		} else {
			err = encoding.Unmarshal(data, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
