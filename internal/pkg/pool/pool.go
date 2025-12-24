package pool

import "sync"

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	syncPool sync.Pool
}

func New[T Resetter](creator func() T) *Pool[T] {
	return &Pool[T]{
		syncPool: sync.Pool{
			New: func() any {
				return creator()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.syncPool.Get().(T)
}

func (p *Pool[T]) Put(value T) {
	value.Reset()
	p.syncPool.Put(value)
}
