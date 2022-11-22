package safe_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aura-studio/lambda/boost/safe"

	. "github.com/frankban/quicktest"
)

func TestDo(t *testing.T) {
	c := New(t)

	c.Run("Do", func(c *C) {
		c.Run("nil", func(c *C) {
			err := safe.Do(nil)
			c.Assert(err, ErrorMatches, "panic: invalid function type: <nil>")
		})

		c.Run("func()", func(c *C) {
			err := safe.Do(func() {})
			c.Assert(err, IsNil)
		})

		c.Run("func() error", func(c *C) {
			err := safe.Do(func() error { return nil })
			c.Assert(err, IsNil)
		})
	})
}

func TestDoWithContext(t *testing.T) {
	c := New(t)

	c.Run("DoWithContext", func(c *C) {
		c.Run("nil", func(c *C) {
			err := safe.DoWithContext(context.Background(), nil)
			c.Assert(err, ErrorMatches, "panic: invalid function type: <nil>")
		})

		c.Run("func()", func(c *C) {
			err := safe.DoWithContext(context.Background(), func() {})
			c.Assert(err, IsNil)
		})

		c.Run("func() error", func(c *C) {
			err := safe.DoWithContext(context.Background(), func() error { return nil })
			c.Assert(err, IsNil)
		})
	})
}

func TestDoWithTimeout(t *testing.T) {
	c := New(t)

	c.Run("DoWithTimeout", func(c *C) {
		c.Run("nil", func(c *C) {
			err := safe.DoWithTimeout(time.Second, nil)
			c.Assert(err, ErrorMatches, "panic: invalid function type: <nil>")
		})

		c.Run("func()", func(c *C) {
			err := safe.DoWithTimeout(time.Second, func() {})
			c.Assert(err, IsNil)
		})

		c.Run("func() error", func(c *C) {
			err := safe.DoWithTimeout(time.Second, func() error { return nil })
			c.Assert(err, IsNil)
		})

		c.Run("timeout", func(c *C) {
			err := safe.DoWithTimeout(time.Nanosecond, func() {
				time.Sleep(time.Second)
			})
			c.Assert(err, ErrorMatches, "context deadline exceeded")
		})

		c.Run("panic", func(c *C) {
			err := safe.DoWithTimeout(time.Second, func() {
				panic("foo")
			})
			c.Assert(err, ErrorMatches, "panic: foo")
		})

		c.Run("error", func(c *C) {
			err := safe.DoWithTimeout(time.Second, func() error {
				return errors.New("foo")
			})
			c.Assert(err, ErrorMatches, "foo")
		})
	})
}
