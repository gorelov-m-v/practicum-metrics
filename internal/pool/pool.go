// Package pool provides a generic object pool that resets objects before reuse.
package pool

import "sync"

// Resetter describes types that can reset their state to zero values.
type Resetter interface {
	Reset()
}

// Pool is a generic wrapper around sync.Pool for objects implementing Resetter.
type Pool[T Resetter] struct {
	p    sync.Pool
	newF func() T
}

// New creates a Pool. The constructor fn must return a ready-to-use object.
func New[T Resetter](fn func() T) *Pool[T] {
	return &Pool[T]{
		newF: fn,
		p: sync.Pool{
			New: func() any { return fn() },
		},
	}
}

// Get returns an object from the pool or creates a new one.
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put resets the object and returns it to the pool.
func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.p.Put(v)
}
