package session

import "sync"

type Registry struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{locks: make(map[string]*sync.Mutex)}
}

func (r *Registry) Get(id string) *sync.Mutex {
	r.mu.Lock()
	defer r.mu.Unlock()
	if mu, ok := r.locks[id]; ok {
		return mu
	}
	mu := &sync.Mutex{}
	r.locks[id] = mu
	return mu
}

func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.locks, id)
}
