// Package doublejump provides a revamped Google's jump consistent hash.
package doublejump

import (
	"math/rand"

	"github.com/dgryski/go-jump"
)

type optional[T comparable] struct {
	b bool
	v T
}

type looseHolder[T comparable] struct {
	a []optional[T]
	m map[T]int
	f []int
}

func (holder *looseHolder[T]) add(obj T) {
	if _, ok := holder.m[obj]; ok {
		return
	}

	if n := len(holder.f); n == 0 {
		holder.a = append(holder.a, optional[T]{v: obj, b: true})
		holder.m[obj] = len(holder.a) - 1
	} else {
		idx := holder.f[n-1]
		holder.f = holder.f[:n-1]
		holder.a[idx] = optional[T]{v: obj, b: true}
		holder.m[obj] = idx
	}
}

func (holder *looseHolder[T]) remove(obj T) {
	if idx, ok := holder.m[obj]; ok {
		holder.a[idx] = optional[T]{}
		holder.f = append(holder.f, idx)
		delete(holder.m, obj)
	}
}

func (holder *looseHolder[T]) get(key uint64) (T, bool) {
	var defVal T
	n := len(holder.a)
	if n == 0 {
		return defVal, false
	}

	h := jump.Hash(key, n)
	if holder.a[h].b {
		return holder.a[h].v, true
	} else {
		return defVal, false
	}
}

func (holder *looseHolder[T]) shrink() {
	if len(holder.f) == 0 {
		return
	}

	var a []optional[T]
	for _, opt := range holder.a {
		if opt.b {
			a = append(a, opt)
			holder.m[opt.v] = len(a) - 1
		}
	}
	holder.a = a
	holder.f = nil
}

type compactHolder[T comparable] struct {
	a []T
	m map[T]int
}

func (holder *compactHolder[T]) add(obj T) {
	if _, ok := holder.m[obj]; ok {
		return
	}

	holder.a = append(holder.a, obj)
	holder.m[obj] = len(holder.a) - 1
}

func (holder *compactHolder[T]) remove(obj T) {
	if idx, ok := holder.m[obj]; ok {
		newLen := len(holder.a) - 1
		tail := holder.a[newLen]
		holder.a[idx] = tail
		holder.m[tail] = idx
		var defVal T
		holder.a[newLen] = defVal
		holder.a = holder.a[:newLen]
		delete(holder.m, obj)
	}
}

func (holder *compactHolder[T]) get(key uint64) (T, bool) {
	var defVal T
	n := len(holder.a)
	if n == 0 {
		return defVal, false
	}

	h := jump.Hash(key*0xc6a4a7935bd1e995, n)
	return holder.a[h], true
}

// Hash is a revamped Google's jump consistent hash. It overcomes the shortcoming of
// the original implementation - not being able to remove nodes.
type Hash[T comparable] struct {
	loose   looseHolder[T]
	compact compactHolder[T]
}

// NewHash creates a new doublejump hash instance, which is NOT thread-safe.
func NewHash[T comparable]() *Hash[T] {
	hash := &Hash[T]{}
	hash.loose.m = make(map[T]int)
	hash.compact.m = make(map[T]int)
	return hash
}

// Add adds an object to the hash.
func (h *Hash[T]) Add(obj T) {
	h.loose.add(obj)
	h.compact.add(obj)
}

// Remove removes an object from the hash.
func (h *Hash[T]) Remove(obj T) {
	h.loose.remove(obj)
	h.compact.remove(obj)
}

// Len returns the number of objects in the hash.
func (h *Hash[T]) Len() int {
	return len(h.compact.a)
}

// LooseLen returns the size of the inner loose object holder.
func (h *Hash[T]) LooseLen() int {
	return len(h.loose.a)
}

// Shrink removes all empty slots from the hash.
func (h *Hash[T]) Shrink() {
	h.loose.shrink()
}

// Get returns an object and a boolean value according to the key provided.
// If there is no object in the hash, ok is false.
func (h *Hash[T]) Get(key uint64) (obj T, ok bool) {
	if obj, ok = h.loose.get(key); ok {
		return obj, true
	}
	return h.compact.get(key)
}

// All returns all the objects in this Hash.
func (h *Hash[T]) All() []T {
	n := len(h.compact.a)
	if n == 0 {
		return nil
	}
	all := make([]T, n)
	copy(all, h.compact.a)
	return all
}

// Random returns a random object.
// If there is no object in the hash, ok is false.
func (h *Hash[T]) Random() (obj T, ok bool) {
	n := len(h.compact.a)
	if n > 0 {
		return h.compact.a[rand.Intn(n)], true
	}
	return *new(T), false
}
