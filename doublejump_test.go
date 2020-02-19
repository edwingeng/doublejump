package doublejump

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"testing/quick"
)

var debugMode = flag.Bool("debug", false, "enable the debug mode")

func always(h *Hash, t *testing.T) {
	if len(h.loose.a) != len(h.loose.m)+len(h.loose.f) {
		t.Fatalf("len(h.loose.a) != len(h.loose.m) + len(h.loose.f). len(a): %d, len(m): %d, len(f): %d",
			len(h.loose.a), len(h.loose.m), len(h.loose.f))
	}
	if len(h.compact.a) != len(h.compact.m) {
		t.Fatalf("len(h.compact.a) != len(h.compact.m). len(a): %d, len(m): %d",
			len(h.compact.a), len(h.compact.m))
	}

	n1 := 0
	for _, obj := range h.loose.a {
		if obj == nil {
			n1++
		}
	}
	if n1 != len(h.loose.f) {
		t.Fatalf("n1 != len(h.loose.f). n1: %d, len(f): %d", n1, len(h.loose.f))
	}

	m1 := make(map[interface{}]int)
	for i, obj := range h.loose.a {
		if obj != nil {
			m1[obj] = i
		}
	}
	if len(m1) != len(h.loose.m) {
		t.Fatalf("len(m1) != len(h.loose.m). len(m1): %d, len(m): %d", len(m1), len(h.loose.m))
	}
	for obj, idx := range h.loose.m {
		if i, ok := m1[obj]; !ok {
			t.Fatalf("cannot find %d in m1", obj)
		} else if i != idx {
			t.Fatalf("m1[%d] != h.loose.m[%d]. idx: %d, i: %d", obj, obj, idx, i)
		}
	}

	m2 := make(map[interface{}]int)
	for i, obj := range h.compact.a {
		if obj != nil {
			m2[obj] = i
		}
	}
	if len(m2) != len(h.compact.m) {
		t.Fatalf("len(m2) != len(h.compact.m). len(m2): %d, len(m): %d", len(m2), len(h.compact.m))
	}
	for obj, idx := range h.compact.m {
		if i, ok := m2[obj]; !ok {
			t.Fatalf("cannot find %d in m2", obj)
		} else if i != idx {
			t.Fatalf("m2[%d] != h.compact.m[%d]. idx: %d, i: %d", obj, obj, idx, i)
		}
	}

	all := h.All()
	if len(all) != h.Len() {
		t.Fatal("len(all) != h.Len()")
	}
}

func TestHash_Basic(t *testing.T) {
	h1 := NewHash()
	always(h1, t)

	var a1 []int
	var n1 = 10
	for i := 0; i < 100*n1; i += 100 {
		h1.Add(i)
		always(h1, t)

		a1 = append(a1, i)
		for loop := 0; loop < 100; loop++ {
			err := quick.Check(func(x int) bool {
				found := false
				obj := h1.Get(uint64(x)).(int)
				for _, v := range a1 {
					if v == obj {
						found = true
						break
					}
				}
				return found
			}, nil)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	h1.Remove(0)
	always(h1, t)

	h1.Remove(100)
	always(h1, t)

	h1.Remove(900)
	always(h1, t)

	h1.Remove(500)
	always(h1, t)

	h1.Shrink()
	always(h1, t)
	h1.Shrink()
	always(h1, t)

	for i := 0; i < 100*n1; i += 100 {
		h1.Remove(i)
		always(h1, t)
	}
}

func TestHash_Add(t *testing.T) {
	h := NewHash()
	h.Add(100)
	h.Add(200)
	h.Add(300)
	h.Add(100)
	always(h, t)

	if h.Len() != 3 {
		t.Fatalf("h.Len() != 3")
	}

	h.Remove(200)
	if h.Len() != 2 {
		t.Fatalf("h.Len() != 2")
	}

	h.Add(500)
	always(h, t)
	if len(h.loose.a) != 3 || h.loose.a[0].(int) != 100 || h.loose.a[1].(int) != 500 || h.loose.a[2].(int) != 300 {
		t.Fatalf("h.loose.a is wrong. a: %v", h.loose.a)
	}
}

func TestHash_Get(t *testing.T) {
	h := NewHash()
	if h.Get(100) != nil {
		t.Fatal("Get should return nil when the hash has no node at all")
	}

	for i := 0; i < 10; i++ {
		h.Add(i)
	}
	for i := 9; i >= 0; i-- {
		h.Remove(i)
	}
	if h.Get(100) != nil {
		t.Fatal("something is wrong with Get")
	}
}

func TestHash_LooseLen(t *testing.T) {
	h := NewHash()
	for i := 0; i < 10; i++ {
		h.Add(i)
	}
	if h.LooseLen() != 10 {
		t.Fatal("h.LooseLen() != 10")
	}

	n := 10
	for i := 1; i < 10; i += 2 {
		h.Remove(i)
		n--
		if h.Len() != n {
			t.Fatalf("h.Len() != n. h.Len(): %d, n: %d", h.Len(), n)
		}
		if h.LooseLen() != 10 {
			t.Fatal("h.LooseLen() should not change after calling Remove")
		}
	}
}

func balance(total uint64, h *Hash, t *testing.T) float64 {
	if h.Len() == 0 {
		return 0
	}
	if total < uint64(h.Len())*10000 {
		panic("total is too small")
	}

	a1 := make([]int, h.LooseLen())
	for i := uint64(0); i < total; i++ {
		a1[h.Get(i).(int)]++
	}
	var nn int
	for _, c := range a1 {
		if c > 0 {
			nn++
		}
	}
	if nn != h.Len() {
		t.Fatalf("nn != h.Len(). nn: %d, h.Len(): %d", nn, h.Len())
	}

	avg := float64(total) / float64(h.Len())
	maxErr := float64(0)
	for obj, c := range a1 {
		if c == 0 {
			continue
		}
		e := math.Abs(float64(c)/avg - 1)
		maxErr = math.Max(maxErr, e)
		if e > 0.15 {
			t.Fatalf("not balance. len: %d, len(f): %d, avg: %.1f, e: %.2f, obj: %c, c: %d",
				h.Len(), len(h.loose.f), avg, e, obj, c)
			break
		}
	}

	return maxErr
}

func TestHash_Balance(t *testing.T) {
	sm := make(chan int, runtime.NumCPU()/2+1)
	for i := 0; i < cap(sm); i++ {
		sm <- 1
	}

	var n1 = 1000
	var wg sync.WaitGroup
	for numRemove := 0; numRemove <= n1; numRemove += rand.Intn(100) {
		<-sm
		wg.Add(1)
		go func(numRemove int) {
			defer func() {
				wg.Done()
				sm <- 1
			}()

			h1 := NewHash()
			a1 := make([]int, 0, n1)
			for i := 0; i < n1; i++ {
				h1.Add(i)
				a1 = append(a1, i)
			}
			rand.Shuffle(n1, func(i, j int) {
				a1[i], a1[j] = a1[j], a1[i]
			})
			for i := 0; i < numRemove; i++ {
				h1.Remove(a1[i])
				always(h1, t)
			}
			if h1.Len() != n1-numRemove {
				t.Fatalf("h1.Len() != n1-numRemove. h1.Len(): %d, n1: %d, numRemove: %d", h1.Len(), n1, numRemove)
			}

			if numRemove < n1 {
				err := quick.Check(func(x int) bool {
					obj := h1.Get(uint64(x))
					if _, ok := h1.loose.m[obj]; !ok {
						return false
					}
					if _, ok := h1.compact.m[obj]; !ok {
						return false
					}
					return true
				}, nil)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				err := quick.Check(func(x int) bool {
					obj := h1.Get(uint64(x))
					return obj == nil
				}, nil)
				if err != nil {
					t.Fatal(err)
				}
			}

			if *debugMode {
				total := uint64(h1.Len()) * 10000
				fmt.Printf("numRemove: %-6d len: %-6d total: %-12d maxErr: %.2f\n",
					numRemove, h1.Len(), total, balance(total, h1, t))
			}
		}(numRemove)
	}

	wg.Wait()
}

func bucket(total uint64, h *Hash) map[interface{}][]uint64 {
	if h.Len() == 0 {
		panic("h.Len() == 0")
	}
	a1 := make([][]uint64, h.LooseLen())
	for i := uint64(0); i < total; i++ {
		obj := h.Get(i).(int)
		a1[obj] = append(a1[obj], i)
	}
	m := make(map[interface{}][]uint64)
	for obj, a2 := range a1 {
		if len(a2) > 0 {
			m[obj] = a2
		}
	}
	return m
}

func TestHash_Consistent(t *testing.T) {
	var n1 = 100
	for step := 1; step < 50; step *= 2 {
		var m0 map[interface{}][]uint64
		for numRemove := 0; numRemove < n1 && numRemove < step*10; numRemove += step {
			h1 := NewHash()
			var a1 []int
			for i := 0; i < n1; i++ {
				h1.Add(i)
				a1 = append(a1, i)
			}
			rand.Shuffle(n1, func(i, j int) {
				a1[i], a1[j] = a1[j], a1[i]
			})
			for i := 0; i < numRemove; i++ {
				h1.Remove(a1[i])
				always(h1, t)
			}
			if h1.Len() != n1-numRemove {
				t.Fatalf("h1.Len() != n1-numRemove. h1.Len(): %d, n1: %d, numRemove: %d", h1.Len(), n1, numRemove)
			}

			total := uint64(n1) * 10000
			m1 := bucket(total, h1)
			if m0 != nil {
				for obj1, a1 := range m1 {
					var a0 = m0[obj1]
					var i0, i1 int
					for i0 < len(a0) && i1 < len(a1) {
						if a1[i1] == a0[i0] {
							i0++
							i1++
						} else {
							i1++
						}
					}
					if i0 < len(a0) {
						t.Fatalf("cannot find %d in a1. i0: %d, numRemove: %d, step: %d\n",
							a0[i0], i0, numRemove, step)
					}
				}
				if *debugMode {
					fmt.Printf("step: %-4d numRemove: %-6d len: %-6d total: %-12d [ok]\n",
						step, numRemove, h1.Len(), total)
				}
			} else {
				m0 = m1
			}
		}
	}
}

func Example() {
	h := NewHash()
	for i := 0; i < 10; i++ {
		h.Add(fmt.Sprintf("node%d", i))
	}

	fmt.Println(h.Len())
	fmt.Println(h.LooseLen())

	fmt.Println(h.Get(1000))
	fmt.Println(h.Get(2000))
	fmt.Println(h.Get(3000))

	h.Remove("node3")
	fmt.Println(h.Len())
	fmt.Println(h.LooseLen())

	fmt.Println(h.Get(1000))
	fmt.Println(h.Get(2000))
	fmt.Println(h.Get(3000))

	// Output:
	// 10
	// 10
	// node9
	// node2
	// node3
	// 9
	// 10
	// node9
	// node2
	// node0
}
