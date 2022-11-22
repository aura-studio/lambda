package message

import (
	"github.com/aura-studio/lambda/core/encoding"
	"github.com/aura-studio/lambda/core/route"
)

type Message struct {
	ID       uint64
	Route    route.Route
	Encoding encoding.Encoding
	Data     []byte
}
