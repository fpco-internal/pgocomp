package pgocomp

import (
	"sync"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var cache map[string]interface{} = make(map[string]interface{})

type Component[T any] struct {
	name  string
	lock  *sync.Mutex
	apply func(ctx *pulumi.Context) (T, error)
}

func NewPulumiComponent[R any, A any, O pulumi.ResourceOption](
	fn func(ctx *pulumi.Context, name string, args A, opts ...O) (R, error),
	name string,
	args A,
	opts ...O,
) *Component[R] {
	return NewComponent(name, func(ctx *pulumi.Context) (R, error) {
		return fn(ctx, name, args, opts...)
	})
}

func NewComponent[T any](name string, apply func(ctx *pulumi.Context) (T, error)) *Component[T] {
	return &Component[T]{name: name, apply: apply, lock: &sync.Mutex{}}
}

func (c *Component[T]) GetAndThen(ctx *pulumi.Context, fn func(T) error) error {
	comp, err := c.Get(ctx)
	if err != nil {
		return err
	}
	return fn(comp)
}

func (c *Component[T]) Apply(ctx *pulumi.Context) error {
	_, err := c.Get(ctx)
	return err
}

func (c *Component[T]) Get(ctx *pulumi.Context) (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	result, ok := cache[c.name]
	if ok {
		return result.(T), nil
	}
	comp, err := c.apply(ctx)
	if err != nil {
		return comp, err
	}
	cache[c.name] = comp
	return comp, nil
}

type Applier interface {
	Apply(ctx *pulumi.Context) error
}

func ApplyAll(ctx *pulumi.Context, components ...Applier) error {
	for _, c := range components {
		if err := c.Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
