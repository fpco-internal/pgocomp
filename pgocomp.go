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
	apply          func(ctx *pulumi.Context, name string) (T, error)
}

// Meta Provides more information to the component
type Meta struct {
	Inactive bool
	Name     string
	Tags     map[string]string
	Protect  bool
}

// FullName is a composition of the NamePrefix and the Name
func (m *Meta) FullName() string {
	if m.Name == "" {
		panic("name and name prefix are blank")
	}
	return m.Name
}

// ComponentWithMeta is a component created using a meta struct
type ComponentWithMeta[T any] struct {
	*Meta
	*Component[T]
}

// NewPulumiComponentWithMeta is a generic function that takes a pulumi "New" function, and all its parameters and returns a component
func NewPulumiComponentWithMeta[R pulumi.Resource, A any, O pulumi.ResourceOption](
	fn func(ctx *pulumi.Context, name string, args A, opts ...O) (R, error),
	meta Meta,
	args A,
	opts ...O,
) *ComponentWithMeta[R] {
	return NewComponentWithMeta[R](meta, func(ctx *pulumi.Context, name string) (R, error) {
		r, err := fn(ctx, name, args, opts...)
		if err == nil {
			ctx.Export(name+"-urn", r.URN())
		}
		return r, err
	})
}

// NewComponentWithMeta is a generic function that takes a name and an apply function and returns a Component
func NewComponentWithMeta[T any](meta Meta, apply func(ctx *pulumi.Context, name string) (T, error)) *ComponentWithMeta[T] {
	return &ComponentWithMeta[T]{
		Meta: &meta,
		Component: func() *Component[T] {
			if meta.Inactive {
				return NewInactiveComponent[T](meta.FullName())
			}
			return NewComponent(meta.FullName(), apply)
		}(),
	}
}

// GetAndThen takes a pulumi context and a function that takes a generic item,
// gets the internal component and
// apply the received function with its internal component
// Can be called multiple times, because it is cached and the component is created only once
func (c *ComponentWithMeta[T]) GetAndThen(ctx *pulumi.Context, fn func(*GetComponentWithMetaResponse[T]) error) error {
	//Do not call underline function when inactive
	if c.Inactive {
		return nil
	}
	comp, err := c.Get(ctx)
	if err != nil {
		return err
	}
	return fn(&GetComponentWithMetaResponse[T]{
		Meta:                 c.Meta,
		GetComponentResponse: comp,
	})
}

// Apply takes a pulumi context, gets the internal component and return error if any. The internal component is discarded.
// Can be called multiple times, because it is cached and the component is created only once
func (c *ComponentWithMeta[T]) Apply(ctx *pulumi.Context) error {
	//Do not call underline function when inactive
	if c.Inactive {
		return nil
	}
	_, err := c.Get(ctx)
	return err
}

// NewPulumiComponent is a generic function that takes a pulumi "New" function, and all its parameters and returns a component
func NewPulumiComponent[R any, A any, O pulumi.ResourceOption](
	fn func(ctx *pulumi.Context, name string, args A, opts ...O) (R, error),
	name string,
	args A,
	opts ...O,
) *Component[R] {
	return NewComponent(name, func(ctx *pulumi.Context, name string) (R, error) {
		return fn(ctx, name, args, opts...)
	})
}

// GetComponentResponse collect the name and the object of created infrastructure components
type GetComponentResponse[T any] struct {
	Name      string
	Component T
}

// GetComponentWithMetaResponse collect the name and the object of created infrastructure components
type GetComponentWithMetaResponse[T any] struct {
	*Meta
	*GetComponentResponse[T]
}

// NewComponent is a generic function that takes a name and an apply function and returns a Component
func NewComponent[T any](name string, apply func(ctx *pulumi.Context, name string) (T, error)) *Component[T] {
	return &Component[T]{name: name, apply: apply, lock: &sync.Mutex{}, isInstantiated: false}
}

// NewInactiveComponent returns a nil internal component
func NewInactiveComponent[T any](name string) *Component[T] {
	return &Component[T]{name: name, apply: func(ctx *pulumi.Context, name string) (T, error) {
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
		if c.element, err = c.apply(ctx, c.name); err != nil {
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

// ExportURNWithMeta exports the URN of pulumi Resource
func ExportURNWithMeta[T pulumi.Resource](ctx *pulumi.Context, name string, r T) {
	ctx.Export(name, r.URN())

}
