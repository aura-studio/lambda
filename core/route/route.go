package route

import (
	"fmt"
	"strings"

	"github.com/aura-studio/lambda/boost/magic"
)

type Route interface {
	String() string
	Position() string
	Dispatching() bool
	Forward() Route
	Reverse() Route
	Error(error) error
}

type ChainRoute struct {
	src   []string
	dst   []string
	index int
}

func NewChainRoute(src, dst []string) *ChainRoute {
	return &ChainRoute{
		src: src,
		dst: dst,
	}
}

func (r ChainRoute) String() string {
	var builder strings.Builder
	builder.WriteString(magic.SeparatorBracketLeft)
	for index, name := range r.src {
		builder.WriteString(name)
		if index != len(r.src)-1 {
			builder.WriteString(magic.SeparatorColon)
		}
	}
	builder.WriteString(magic.SeparatorBracketRight)
	builder.WriteString(magic.SeparatorSpace)
	builder.WriteString(magic.SeparatorMinus)
	builder.WriteString(magic.SeparatorGreater)
	builder.WriteString(magic.SeparatorSpace)
	builder.WriteString(magic.SeparatorBracketLeft)
	for index, name := range r.dst {
		if index == r.index {
			builder.WriteString(magic.SeparatorLess)
			builder.WriteString(name)
			builder.WriteString(magic.SeparatorGreater)
		} else {
			builder.WriteString(name)
		}
		if index != len(r.dst)-1 {
			builder.WriteString(magic.SeparatorColon)
		}
	}
	builder.WriteString(magic.SeparatorBracketRight)
	return builder.String()
}

func (r ChainRoute) Dispatching() bool {
	return r.index > 0
}

func (r ChainRoute) Forward() Route {
	if r.index < len(r.dst)-1 {
		r.index++
	}
	return r
}

func (r ChainRoute) Position() string {
	return r.dst[r.index]
}

func (r ChainRoute) Reverse() Route {
	return ChainRoute{
		src:   r.dst,
		dst:   r.src,
		index: 0,
	}
}

func (r ChainRoute) Error(err error) error {
	return fmt.Errorf("route %v error: %w", r, err)
}
