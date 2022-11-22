package safe

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aura-studio/lambda/boost/cast"
)

func Do(a any) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("panic: %w", cast.ToError(v))
		}
	}()

	switch f := a.(type) {
	case func():
		f()
	case func() error:
		err = f()
	default:
		panic(fmt.Errorf("invalid function type: %T", f))
	}

	return
}

func DoWithContext(ctx context.Context, a any) (err error) {
	errCh := make(chan error, 1)

	go func() {
		errCh <- Do(a)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errCh:
		return err
	}
}

func DoWithTimeout(d time.Duration, a any) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	return DoWithContext(ctx, a)
}

type Lock struct {
	sync.Mutex
}

func (l *Lock) Do(a any) (err error) {
	l.Lock()
	defer l.Unlock()

	return Do(a)
}

func (l *Lock) DoWithContext(ctx context.Context, a any) (err error) {
	l.Lock()
	defer l.Unlock()

	return DoWithContext(ctx, a)
}

func (l *Lock) DoWithTimeout(d time.Duration, a any) (err error) {
	l.Lock()
	defer l.Unlock()

	return DoWithTimeout(d, a)
}

type RWLock struct {
	sync.RWMutex
}

func (l *RWLock) Do(a any) (err error) {
	l.Lock()
	defer l.Unlock()

	return Do(a)
}

func (l *RWLock) DoWithContext(ctx context.Context, a any) (err error) {
	l.Lock()
	defer l.Unlock()

	return DoWithContext(ctx, a)
}

func (l *RWLock) DoWithTimeout(d time.Duration, a any) (err error) {
	l.Lock()
	defer l.Unlock()

	return DoWithTimeout(d, a)
}

func (l *RWLock) RDo(a any) (err error) {
	l.RLock()
	defer l.RUnlock()

	return Do(a)
}

func (l *RWLock) RDoWithContext(ctx context.Context, a any) (err error) {
	l.RLock()
	defer l.RUnlock()

	return DoWithContext(ctx, a)
}

func (l *RWLock) RDoWithTimeout(d time.Duration, a any) (err error) {
	l.RLock()
	defer l.RUnlock()

	return DoWithTimeout(d, a)
}
