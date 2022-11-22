package encoding_test

import (
	"reflect"
	"testing"

	"github.com/aura-studio/lambda/boost/magic"
	"github.com/aura-studio/lambda/core/encoding"
)

func TestChain(t *testing.T) {
	type TestStruct struct {
		Integer int
		String  string
		Bool    bool
		Slice   []byte
	}

	ts1 := &TestStruct{
		Integer: 1,
		String:  "this is test text",
		Bool:    false,
		Slice:   []byte("this is slice"),
	}
	t.Logf("ts1: %+v", ts1)
	e1 := encoding.NewChainEncoding(magic.GoogleChain("json.base64.lazy"), magic.GoogleChain("lazy.base64.xml"))
	t.Logf("e1: %v", e1)

	e2 := e1.Reverse()
	t.Logf("e2: %v", e2)

	data1, err := encoding.Marshal(e1, ts1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("data1: %s", string(data1))

	ts2 := &TestStruct{}
	err = encoding.Unmarshal(e2, data1, ts2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ts1, ts2) {
		t.Fatalf("expecting ts1 %v equals ts2 %v", ts1, ts2)
	}
	t.Logf("ts2: %+v", ts2)

	data2, err := encoding.Marshal(e2, ts2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("data2: %s", string(data2))

	ts3 := &TestStruct{}
	err = encoding.Unmarshal(e1, data2, ts3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ts2, ts3) {
		t.Fatalf("expecting ts2 %v equals ts3 %v", ts2, ts3)
	}
	t.Logf("ts3: %+v", ts3)
}
