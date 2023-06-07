package pgocomp

import (
	"sync"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Component is a generic struct that allows for a diffent way to write Lazy and Idempotent Pulumi components.
type Component[T any] struct {
	name           string
	lock           *sync.Mutex
	isInstantiated bool
	element        T
	apply          func(ctx *pulumi.Context) (T, error)
}

// NewLazyArgsPulumiComponent is a generic function that takes a pulumi "New" function, a function that returns its args and options and returns a component
func NewLazyArgsPulumiComponent[R any, A any, O pulumi.ResourceOption](
	fn func(ctx *pulumi.Context, name string, args A, opts ...O) (R, error),
	name string,
	argsFn func(ctx *pulumi.Context) (A, []O, error),
) *Component[R] {
	return NewComponent(name, func(ctx *pulumi.Context) (R, error) {
		args, opts, err := argsFn(ctx)
		if err != nil {
			var result R
			return result, err
		}
		return fn(ctx, name, args, opts...)
	})
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

// GetComponentResponse collect the name and the object of created infrastructure components
type GetComponentResponse[T any] struct {
	Name      string
	Component T
}

// NewComponent is a generic function that takes a name and an apply function and returns a Component
func NewComponent[T any](name string, apply func(ctx *pulumi.Context) (T, error)) *Component[T] {
	return &Component[T]{name: name, apply: apply, lock: &sync.Mutex{}, isInstantiated: false}
}

// NewInactiveComponent returns a nil internal component
func NewInactiveComponent[T any](name string) *Component[T] {
	return &Component[T]{name: name, apply: func(ctx *pulumi.Context) (T, error) {
		var t T
		return t, nil
	}, lock: &sync.Mutex{}, isInstantiated: false}
}

// GetAndThen takes a pulumi context and a function that takes a generic item,
// gets the internal component and
// apply the received function with its internal component
// Can be called multiple times, because it is cached and the component is created only once
func (c *Component[T]) GetAndThen(ctx *pulumi.Context, fn func(*GetComponentResponse[T]) error) error {
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
func (c *Component[T]) Get(ctx *pulumi.Context) (*GetComponentResponse[T], error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.isInstantiated {
		var err error
		if c.element, err = c.apply(ctx); err != nil {
			return &GetComponentResponse[T]{Name: c.name, Component: c.element}, err
		}
		c.isInstantiated = true
	}
	return &GetComponentResponse[T]{Name: c.name, Component: c.element}, nil
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

// ExportURN exports the URN of pulumi Resource
func ExportURN[T pulumi.Resource](ctx *pulumi.Context, r *GetComponentResponse[T]) *GetComponentResponse[T] {
	ctx.Export(r.Name+"-id", r.Component.URN())
	return r
}
