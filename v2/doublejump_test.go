package doublejump

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

var debugMode = flag.Bool("debugMode", false, "enable the debug mode")

func init() {
	rand.Seed(time.Now().UnixMilli())
}

func invariant[T comparable](h *Hash[T], t *testing.T) {
	t.Helper()
	if len(h.loose.a) != len(h.loose.m)+len(h.loose.f) {
		t.Fatalf("len(h.loose.a) != len(h.loose.m) + len(h.loose.f). len(a): %d, len(m): %d, len(f): %d",
			len(h.loose.a), len(h.loose.m), len(h.loose.f))
	}
	if len(h.compact.a) != len(h.compact.m) {
		t.Fatalf("len(h.compact.a) != len(h.compact.m). len(a): %d, len(m): %d",
			len(h.compact.a), len(h.compact.m))
	}

	for obj, idx := range h.loose.m {
		if opt := h.loose.a[idx]; !opt.b || opt.v != obj {
			t.Fatalf(`!opt.b || opt.v != obj. obj: %v, idx: %v, opt.b: %v, opt.v: %v`, obj, idx, opt.b, opt.v)
		}
	}

	var defVal T
	freeMap := make(map[int]struct{})
	m1 := make(map[int]struct{})
	for _, idx := range h.loose.f {
		freeMap[idx] = struct{}{}
		m1[idx] = struct{}{}
		if h.loose.a[idx].v != defVal {
			t.Fatalf(`h.loose.a[idx].v != defVal. idx: %d, a[idx]: %v`, idx, h.loose.a[idx])
		}
	}
	if len(freeMap) != len(h.loose.f) {
		t.Fatalf("len(freeMap) != len(h.loose.f). %d vs %d",
			len(freeMap), len(h.loose.f))
	}

	slots := make([]bool, len(h.loose.a))
	usedMap := make(map[int]struct{})
	for _, idx := range h.loose.m {
		slots[idx] = true
		usedMap[idx] = struct{}{}
		m1[idx] = struct{}{}
		if _, ok := freeMap[idx]; ok {
			t.Fatalf("%d should not be in the free list", idx)
		}
	}
	for i := range slots {
		if h.loose.a[i].b != slots[i] {
			t.Fatalf("h.loose.a[i].b != slots[i]. i: %d, b[i]: %v, slots[i]: %v",
				i, h.loose.a[i].b, slots[i])
		}
	}
	if len(usedMap) != len(h.loose.m) {
		t.Fatalf("len(usedMap) != len(h.loose.m). %d vs %d",
			len(usedMap), len(h.loose.m))
	}
	if len(m1) != len(h.loose.a) {
		t.Fatalf("len(m1) != len(h.loose.a). %d vs %d",
			len(m1), len(h.loose.a))
	}

	m2 := make(map[T]int)
	for i, obj := range h.compact.a {
		m2[obj] = i
	}
	if len(m2) != len(h.compact.m) {
		t.Fatalf("len(m2) != len(h.compact.m). len(m2): %d, len(m): %d", len(m2), len(h.compact.m))
	}
	for obj, idx := range h.compact.m {
		if i, ok := m2[obj]; !ok {
			t.Fatalf("cannot find %v in m2", obj)
		} else if i != idx {
			t.Fatalf("m2[%v] != h.compact.m[%v]. idx: %d, i: %d", obj, obj, idx, i)
		}
	}

	all := h.All()
	if len(all) != h.Len() {
		t.Fatal("len(all) != h.Len()")
	}
}

func TestHash_Basic(t *testing.T) {
	h := NewHash[int]()
	invariant(h, t)

	const n1 = 10
	for i := 0; i < 100*n1; i += 100 {
		h.Add(i)
		invariant(h, t)

		for j := 0; j <= i; j++ {
			if _, ok := h.Get(uint64(j)); !ok {
				t.Fatal("something is wrong with Get")
			}
		}
		for j := 0; j < 10000; j++ {
			if _, ok := h.Get(rand.Uint64()); !ok {
				t.Fatal("something is wrong with Get")
			}
		}
	}

	h.Remove(0)
	invariant(h, t)

	h.Remove(100)
	invariant(h, t)

	h.Remove(900)
	invariant(h, t)

	h.Remove(500)
	invariant(h, t)

	h.Shrink()
	invariant(h, t)
	h.Shrink()
	invariant(h, t)

	for i := 0; i < 100*n1; i += 100 {
		h.Remove(i)
		invariant(h, t)
	}
}

func TestHash_Add(t *testing.T) {
	h := NewHash[int]()
	h.Add(100)
	h.Add(200)
	h.Add(300)
	h.Add(100)
	invariant(h, t)

	if h.Len() != 3 {
		t.Fatalf("h.Len() != 3")
	}

	h.Remove(200)
	if h.Len() != 2 {
		t.Fatalf("h.Len() != 2")
	}

	h.Add(500)
	invariant(h, t)
	if len(h.loose.a) != 3 || h.loose.a[0].v != 100 || h.loose.a[1].v != 500 || h.loose.a[2].v != 300 {
		t.Fatalf("h.loose.a is wrong. a: %v", h.loose.a)
	}
}

func TestHash_Get(t *testing.T) {
	h := NewHash[int]()
	if v, ok := h.Get(100); ok || v != 0 {
		t.Fatal("something is wrong with Get")
	}

	for i := 0; i < 10; i++ {
		h.Add(i)
	}
	for i := 9; i >= 0; i-- {
		h.Remove(i)
	}
	if v, ok := h.Get(100); ok || v != 0 {
		t.Fatal("something is wrong with Get")
	}
}

func TestHash_LooseLen(t *testing.T) {
	h := NewHash[int]()
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

func checkBalance(total int, h *Hash[int]) (float64, error) {
	if h.Len() == 0 {
		return 0, nil
	}
	if total < h.Len()*10000 {
		return 0, errors.New("total is too small")
	}

	a := make([]int, h.LooseLen())
	for i := 0; i < total; i++ {
		if v, ok := h.Get(uint64(i)); ok {
			a[v]++
		} else {
			panic("impossible")
		}
	}
	var nn int
	for _, c := range a {
		if c > 0 {
			nn++
		}
	}
	if nn != h.Len() {
		return 0, fmt.Errorf("nn != h.Len(). nn: %d, h.Len(): %d", nn, h.Len())
	}

	maxErr := float64(0)
	avg := float64(total) / float64(h.Len())
	for obj, c := range a {
		if c == 0 {
			continue
		}
		e := math.Abs(float64(c)/avg - 1)
		maxErr = math.Max(maxErr, e)
		if e > 0.15 {
			return 0, fmt.Errorf("not balanced. len: %d, len(f): %d, avg: %.1f, e: %.2f, obj: %c, c: %d",
				h.Len(), len(h.loose.f), avg, e, obj, c)
		}
	}

	return maxErr, nil
}

func TestHash_Balance(t *testing.T) {
	sm := make(chan int, runtime.NumCPU()/2+1)
	for i := 0; i < cap(sm); i++ {
		sm <- 1
	}

	const n1 = 1000
	chErr := make(chan error, n1)
	var wg sync.WaitGroup
	for rm := 0; true; rm += rand.Intn(100) + 1 {
		rm := rm
		if rm > n1 {
			rm = n1
		}

		<-sm
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				sm <- 1
			}()

			h := NewHash[int]()
			var a [n1]int
			for i := 0; i < n1; i++ {
				h.Add(i)
				a[i] = i
			}
			rand.Shuffle(n1, func(i, j int) {
				a[i], a[j] = a[j], a[i]
			})
			for i := 0; i < rm; i++ {
				h.Remove(a[i])
				invariant(h, t)
			}
			if h.Len() != n1-rm {
				chErr <- fmt.Errorf("h.Len() != n1-rm. h.Len(): %d, n1: %d, rm: %d", h.Len(), n1, rm)
				return
			}

			if rm != n1 {
				for j := 0; j < 10000; j++ {
					if _, ok := h.Get(rand.Uint64()); !ok {
						chErr <- fmt.Errorf("something is wrong with Get [1]. len: %d, looseLen: %d",
							h.Len(), h.LooseLen())
						return
					}
				}
			} else {
				if _, ok := h.Get(rand.Uint64()); ok {
					chErr <- fmt.Errorf("something is wrong with Get [2]. len: %d, looseLen: %d",
						h.Len(), h.LooseLen())
					return
				}
			}

			if *debugMode {
				total := h.Len() * 10000
				maxErr, err := checkBalance(total, h)
				if err != nil {
					chErr <- err
					return
				}
				fmt.Printf("rm: %-6d len: %-6d total: %-12d maxErr: %.2f\n",
					rm, h.Len(), total, maxErr)
			}
		}()

		if rm == n1 {
			break
		}
	}

	wg.Wait()
	select {
	case err := <-chErr:
		t.Fatal(err)
	default:
	}
}

func TestHash_Consistent(t *testing.T) {
	const n1 = 100
	numbers := []int{0, 2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47}
	var m0 map[int]int
	for _, rm := range numbers {
		h1 := NewHash[int]()
		var xx []int
		for i := 0; i < n1; i++ {
			h1.Add(i)
			xx = append(xx, i)
		}
		rand.Shuffle(n1, func(i, j int) {
			xx[i], xx[j] = xx[j], xx[i]
		})
		for i := 0; i < rm; i++ {
			h1.Remove(xx[i])
			invariant(h1, t)
		}
		if h1.Len() != n1-rm {
			t.Fatalf("h1.Len() != n1-rm. h1.Len(): %d, n1: %d, rm: %d", h1.Len(), n1, rm)
		}

		total := n1 * 10000
		m1 := make(map[int]int, total)
		for i := 0; i < total; i++ {
			if obj, ok := h1.Get(uint64(i)); ok {
				m1[i] = obj
			} else {
				panic("impossible")
			}
		}

		switch rm {
		case 0:
			m0 = m1
		default:
			var n2 int
			for k, v := range m1 {
				if m0[k] == v {
					n2++
				}
			}
			r1 := float64(total-n2) / float64(total)
			r2 := float64(rm) / n1
			delta := math.Abs(r1 - r2)
			if delta > 0.05 {
				t.Fatal("delta > 0.05")
			}
		}
	}
}

func Example() {
	h := NewHash[string]()
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
	// node9 true
	// node2 true
	// node3 true
	// 9
	// 10
	// node9 true
	// node2 true
	// node0 true
}
