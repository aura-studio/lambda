package encoding

import (
	"bytes"

	"github.com/aura-studio/lambda/boost/ref"

	"github.com/gocarina/gocsv"
)

type CSV struct{}

func init() {
	register(NewCSV())
}

func NewCSV() *CSV {
	return new(CSV)
}

func (csv CSV) String() string {
	return ref.TypeName(csv)
}

func (CSV) Style() EncodingStyleType {
	return EncodingStyleStruct
}

func (CSV) Marshal(v interface{}) ([]byte, error) {
	var buf = new(bytes.Buffer)
	err := gocsv.MarshalWithoutHeaders(v, buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (CSV) Unmarshal(data []byte, v interface{}) error {
	return gocsv.UnmarshalWithoutHeaders(bytes.NewBuffer(data), v)
}

func (csv CSV) Reverse() Encoding {
	return csv
}

type CSVWithHeaders struct{}

func init() {
	register(NewCSVWithHeaders())
}

func NewCSVWithHeaders() *CSVWithHeaders {
	return new(CSVWithHeaders)
}

func (csvwh CSVWithHeaders) String() string {
	return ref.TypeName(csvwh)
}

func (CSVWithHeaders) Style() EncodingStyleType {
	return EncodingStyleStruct
}

func (CSVWithHeaders) Marshal(v interface{}) ([]byte, error) {
	var buf = new(bytes.Buffer)
	err := gocsv.Marshal(v, buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (CSVWithHeaders) Unmarshal(data []byte, v interface{}) error {
	return gocsv.Unmarshal(bytes.NewBuffer(data), v)
}

func (csvwh CSVWithHeaders) Reverse() Encoding {
	return csvwh
}
