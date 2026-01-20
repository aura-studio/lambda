package sqs

import (
	"google.golang.org/protobuf/proto"
)

func MarshalRequest(r *Request) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	return proto.Marshal(r)
}

func UnmarshalRequest(b []byte) (*Request, error) {
	var r Request
	if err := proto.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func MarshalResponse(r *Response) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	return proto.Marshal(r)
}

func UnmarshalResponse(b []byte) (*Response, error) {
	var r Response
	if err := proto.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func DecodeRequestBody(body string) (*Request, error) {
	return UnmarshalRequest([]byte(body))
}
