// Package doublejump provides a revamped Google's jump consistent hash.
package doublejump

import (
	"sync"

	"github.com/dgryski/go-jump"
)

type looseHolder struct {
	a []interface{}
	m map[interface{}]int
	f []int
}

func (xh *looseHolder) add(obj interface{}) {
	if _, ok := xh.m[obj]; ok {
		return
	}

	if nf := len(xh.f); nf == 0 {
		xh.a = append(xh.a, obj)
		xh.m[obj] = len(xh.a) - 1
	} else {
		idx := xh.f[nf-1]
		xh.f = xh.f[:nf-1]
		xh.a[idx] = obj
		xh.m[obj] = idx
	}
}

func (xh *looseHolder) remove(obj interface{}) {
	if idx, ok := xh.m[obj]; ok {
		xh.f = append(xh.f, idx)
		xh.a[idx] = nil
		delete(xh.m, obj)
	}
}

func (xh *looseHolder) get(key uint64) interface{} {
	na := len(xh.a)
	if na == 0 {
		return nil
	}

	h := jump.Hash(key, na)
	return xh.a[h]
}

func (xh *looseHolder) shrink() {
	if len(xh.f) == 0 {
		return
	}

	var a []interface{}
	for _, obj := range xh.a {
		if obj != nil {
			a = append(a, obj)
			xh.m[obj] = len(a) - 1
		}
	}
	xh.a = a
	xh.f = nil
}

type compactHolder struct {
	a []interface{}
	m map[interface{}]int
}

func (xh *compactHolder) add(obj interface{}) {
	if _, ok := xh.m[obj]; ok {
		return
	}

	xh.a = append(xh.a, obj)
	xh.m[obj] = len(xh.a) - 1
}

func (xh *compactHolder) shrink(a []interface{}) {
	for i, obj := range a {
		xh.a[i] = obj
		xh.m[obj] = i
	}
}

func (xh *compactHolder) remove(obj interface{}) {
	if idx, ok := xh.m[obj]; ok {
		n := len(xh.a)
		xh.a[idx] = xh.a[n-1]
		xh.m[xh.a[idx]] = idx
		xh.a[n-1] = nil
		xh.a = xh.a[:n-1]
		delete(xh.m, obj)
	}
}

func (xh *compactHolder) get(key uint64, nl int) interface{} {
	na := len(xh.a)
	if na == 0 {
		return nil
	}

	h := jump.Hash(uint64(float64(key)/float64(nl)*float64(na)), na)
	return xh.a[h]
}

// Hash is a revamped Google's jump consistent hash. It overcomes the shortcoming of the
// original implementation - not being able to remove nodes.
type Hash struct {
	mu      sync.RWMutex
	loose   looseHolder
	compact compactHolder
	lock    bool
}

// NewHash creates a new doublejump hash instance, which is threadsafe.
func NewHash() *Hash {
	hash := &Hash{lock: true}
	hash.loose.m = make(map[interface{}]int)
	hash.compact.m = make(map[interface{}]int)
	return hash
}

// NewHashWithoutLock creates a new doublejump hash instance, which does NOT threadsafe.
func NewHashWithoutLock() *Hash {
	hash := &Hash{}
	hash.loose.m = make(map[interface{}]int)
	hash.compact.m = make(map[interface{}]int)
	return hash
}

// Add adds an object to the hash.
func (xh *Hash) Add(obj interface{}) {
	if xh == nil || obj == nil {
		return
	}

	if xh.lock {
		xh.mu.Lock()
		defer xh.mu.Unlock()
	}

	xh.loose.add(obj)
	xh.compact.add(obj)
}

// Remove removes an object from the hash.
func (xh *Hash) Remove(obj interface{}) {
	if xh == nil || obj == nil {
		return
	}

	if xh.lock {
		xh.mu.Lock()
		defer xh.mu.Unlock()
	}

	xh.loose.remove(obj)
	xh.compact.remove(obj)
}

// Len returns the number of objects in the hash.
func (xh *Hash) Len() int {
	if xh == nil {
		return 0
	}

	if xh.lock {
		xh.mu.RLock()
		n := len(xh.compact.a)
		xh.mu.RUnlock()
		return n
	}

	return len(xh.compact.a)
}

// LooseLen returns the size of the inner loose object holder.
func (xh *Hash) LooseLen() int {
	if xh == nil {
		return 0
	}

	if xh.lock {
		xh.mu.RLock()
		n := len(xh.loose.a)
		xh.mu.RUnlock()
		return n
	}

	return len(xh.loose.a)
}

// Shrink removes all empty slots from the hash.
func (xh *Hash) Shrink() {
	if xh == nil {
		return
	}

	if xh.lock {
		xh.mu.Lock()
		defer xh.mu.Unlock()
	}

	xh.loose.shrink()
	xh.compact.shrink(xh.loose.a)
}

// Get returns an object according to the key provided.
func (xh *Hash) Get(key uint64) interface{} {
	if xh == nil {
		return nil
	}

	if xh.lock {
		xh.mu.RLock()
		defer xh.mu.RUnlock()
	}

	obj := xh.loose.get(key)
	if obj != nil {
		return obj
	}

	return xh.compact.get(key, len(xh.loose.a))
}
