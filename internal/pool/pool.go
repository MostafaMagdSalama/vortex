package pool

import "sync"

// Pool is a generic wrapper around sync.Pool.
// T is the type of object being pooled.
type Pool[T any] struct {
    p sync.Pool
}

// NewPool creates a pool with a factory function that makes new T when empty.
func NewPool[T any](factory func() T) *Pool[T] {
    return &Pool[T]{
        p: sync.Pool{
            New: func() any {
                return factory()
            },
        },
    }
}

// Get retrieves an object from the pool.
// If pool is empty, factory is called to create a new one.
func (p *Pool[T]) Get() T {
    return p.p.Get().(T)
}

// Put returns an object back to the pool for reuse.
func (p *Pool[T]) Put(v T) {
    p.p.Put(v)
}