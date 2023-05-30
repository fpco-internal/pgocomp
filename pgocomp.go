package pgocomp

import (
	"sync"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var cache map[string]interface{} = make(map[string]interface{})

// Component is a generic struct that allows for a diffent way to write Lazy and Idempotent Pulumi components.
type Component[T any] struct {
	name  string
	lock  *sync.Mutex
	apply func(ctx *pulumi.Context) (T, error)
}

// NewPulumiComponent is a generic function that takes a pulumi "New" function, and all its parameters and returns a component
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

// NewComponent is a generic function that takes a name and an apply function and returns a Component
func NewComponent[T any](name string, apply func(ctx *pulumi.Context) (T, error)) *Component[T] {
	return &Component[T]{name: name, apply: apply, lock: &sync.Mutex{}}
}

// GetAndThen takes a pulumi context and a function that takes a generic item,
// gets the internal component and
// apply the received function with its internal component
// Can be called multiple times, because it is cached and the component is created only once
func (c *Component[T]) GetAndThen(ctx *pulumi.Context, fn func(T) error) error {
	comp, err := c.Get(ctx)
	if err != nil {
		return err
	}
	return fn(comp)
}

// Apply takes a pulumi context, gets the internal component and return error if any. The internal component is discarded.
// Can be called multiple times, because it is cached and the component is created only once
func (c *Component[T]) Apply(ctx *pulumi.Context) error {
	_, err := c.Get(ctx)
	return err
}

// Get takes a pulumi context, gets the internal component and return error if any. The internal component is discarded.
// Can be called multiple times, because it is cached and the component is created only once
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

// Applier is an interface created for the ApplyAll function
type Applier interface {
	//Apply is a function that executes using a pulumi context and that returns an error
	Apply(ctx *pulumi.Context) error
}

// ApplyAll is a function that takes a variadic list of appliers and calls the Apply method on each on of them
func ApplyAll(ctx *pulumi.Context, appliers ...Applier) error {
	for _, c := range appliers {
		if err := c.Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
