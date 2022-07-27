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
	rand.Seed(time.Now().UnixNano() / 1000)
}

func invariant(h *Hash, t *testing.T) {
	t.Helper()
	err := invariantImpl(h)
	if err != nil {
		t.Fatal(err)
	}
}

//gocyclo:ignore
func invariantImpl(h *Hash) error {
	if len(h.loose.a) != len(h.loose.m)+len(h.loose.f) {
		return fmt.Errorf("len(h.loose.a) != len(h.loose.m) + len(h.loose.f). len(a): %d, len(m): %d, len(f): %d",
			len(h.loose.a), len(h.loose.m), len(h.loose.f))
	}
	if len(h.compact.a) != len(h.compact.m) {
		return fmt.Errorf("len(h.compact.a) != len(h.compact.m). len(a): %d, len(m): %d",
			len(h.compact.a), len(h.compact.m))
	}

	for obj, idx := range h.loose.m {
		if h.loose.a[idx] != obj {
			return fmt.Errorf(`h.loose.a[idx] != obj. obj: %v, idx: %v, a[idx]: %v`,
				obj, idx, h.loose.a[idx])
		}
	}

	freeMap := make(map[int]struct{})
	m1 := make(map[int]struct{})
	for _, idx := range h.loose.f {
		freeMap[idx] = struct{}{}
		m1[idx] = struct{}{}
		if h.loose.a[idx] != nil {
			return fmt.Errorf(`h.loose.a[idx] != nil. idx: %d, a[idx]: %v`, idx, h.loose.a[idx])
		}
	}
	if len(freeMap) != len(h.loose.f) {
		return fmt.Errorf("len(freeMap) != len(h.loose.f). %d vs %d",
			len(freeMap), len(h.loose.f))
	}

	slots := make([]bool, len(h.loose.a))
	usedMap := make(map[int]struct{})
	for _, idx := range h.loose.m {
		slots[idx] = true
		usedMap[idx] = struct{}{}
		m1[idx] = struct{}{}
		if _, ok := freeMap[idx]; ok {
			return fmt.Errorf("%d should not be in the free list", idx)
		}
	}
	for i := range slots {
		if yes := h.loose.a[i] != nil; yes != slots[i] {
			return fmt.Errorf("yes != slots[i]. i: %d, yes: %v, slots[i]: %v",
				i, yes, slots[i])
		}
	}
	if len(usedMap) != len(h.loose.m) {
		return fmt.Errorf("len(usedMap) != len(h.loose.m). %d vs %d",
			len(usedMap), len(h.loose.m))
	}
	if len(m1) != len(h.loose.a) {
		return fmt.Errorf("len(m1) != len(h.loose.a). %d vs %d",
			len(m1), len(h.loose.a))
	}

	m2 := make(map[interface{}]int)
	for i, obj := range h.compact.a {
		m2[obj] = i
	}
	if len(m2) != len(h.compact.m) {
		return fmt.Errorf("len(m2) != len(h.compact.m). len(m2): %d, len(m): %d", len(m2), len(h.compact.m))
	}
	for obj, idx := range h.compact.m {
		if i, ok := m2[obj]; !ok {
			return fmt.Errorf("cannot find %v in m2", obj)
		} else if i != idx {
			return fmt.Errorf("m2[%v] != h.compact.m[%v]. idx: %d, i: %d", obj, obj, idx, i)
		}
	}

	all := h.All()
	if len(all) != h.Len() {
		return fmt.Errorf("len(all) != h.Len()")
	}

	return nil
}

func TestHash_Basic(t *testing.T) {
	h := NewHash()
	invariant(h, t)

	const n1 = 10
	for i := 0; i < 100*n1; i += 100 {
		h.Add(i)
		invariant(h, t)

		for j := 0; j <= i; j++ {
			if h.Get(uint64(j)) == nil {
				t.Fatal("something is wrong with Get")
			}
		}
		for j := 0; j < 10000; j++ {
			if h.Get(rand.Uint64()) == nil {
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
	h := NewHash()
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
	if len(h.loose.a) != 3 || h.loose.a[0] != 100 || h.loose.a[1] != 500 || h.loose.a[2] != 300 {
		t.Fatalf("h.loose.a is wrong. a: %v", h.loose.a)
	}
}

func TestHash_Get(t *testing.T) {
	h := NewHash()
	if h.Get(100) != nil {
		t.Fatal("something is wrong with Get")
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

func checkBalance(total int, h *Hash) (float64, error) {
	if h.Len() == 0 {
		return 0, nil
	}
	if total < h.Len()*10000 {
		return 0, errors.New("total is too small")
	}

	a := make([]int, h.LooseLen())
	for i := 0; i < total; i++ {
		if v := h.Get(uint64(i)); v != nil {
			a[v.(int)]++
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

//gocyclo:ignore
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

			h := NewHash()
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
				if err := invariantImpl(h); err != nil {
					chErr <- err
					return
				}
			}
			if h.Len() != n1-rm {
				chErr <- fmt.Errorf("h.Len() != n1-rm. h.Len(): %d, n1: %d, rm: %d", h.Len(), n1, rm)
				return
			}

			if rm != n1 {
				for j := 0; j < 10000; j++ {
					if h.Get(rand.Uint64()) == nil {
						chErr <- fmt.Errorf("something is wrong with Get [1]. len: %d, looseLen: %d",
							h.Len(), h.LooseLen())
						return
					}
				}
			} else {
				if h.Get(rand.Uint64()) != nil {
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

//gocyclo:ignore
func TestHash_Consistent(t *testing.T) {
	const nn = 100
	const total = nn * 10000
	run := func(h *Hash) map[int]interface{} {
		m1 := make(map[int]interface{}, total)
		for i := 0; i < total; i++ {
			if obj := h.Get(uint64(i)); obj != nil {
				m1[i] = obj
			} else {
				panic("impossible")
			}
		}
		return m1
	}

	checkResults := func(r1, r2 float64) error {
		delta := r1 - r2
		if delta > 0.05 {
			return fmt.Errorf("delta > 0.05. r1: %.4f, r2: %.4f, delta: %.4f", r1, r2, delta)
		}
		return nil
	}

	h1 := NewHash()
	for i := 0; i < nn; i++ {
		h1.Add(i)
	}

	m0 := run(h1)
	numbers := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47}
	chErr := make(chan error, len(numbers))
	var wg sync.WaitGroup
	wg.Add(len(numbers))
	for _, rm := range numbers {
		rm := rm
		go func() {
			defer wg.Done()
			h2 := NewHash()
			var xx []int
			for i := 0; i < nn; i++ {
				h2.Add(i)
				xx = append(xx, i)
			}
			rand.Shuffle(nn, func(i, j int) {
				xx[i], xx[j] = xx[j], xx[i]
			})
			for i := 0; i < rm; i++ {
				h2.Remove(xx[i])
				if err := invariantImpl(h2); err != nil {
					chErr <- err
					return
				}
			}
			if h2.Len() != nn-rm {
				chErr <- fmt.Errorf("h2.Len() != nn-rm. h2.Len(): %d, nn: %d, rm: %d", h2.Len(), nn, rm)
				return
			}

			m1 := run(h2)
			var n1 int
			for k, v := range m1 {
				if m0[k] == v {
					n1++
				}
			}
			err := checkResults(float64(total-n1)/total, float64(rm)/nn)
			if err != nil {
				chErr <- err
				return
			}

			if c2 := rm / 3; c2 > 0 {
				for i := 0; i < c2; i++ {
					h2.Add(xx[i])
					if err := invariantImpl(h2); err != nil {
						chErr <- err
						return
					}
				}
				m2 := run(h2)
				var n2 int
				for k, v := range m2 {
					if m1[k] == v {
						n2++
					}
				}
				err := checkResults(float64(total-n2)/total, float64(c2)/float64(h2.Len()))
				if err != nil {
					chErr <- err
					return
				}

				if c3 := rm / 7; c3 > 0 {
					for i := 0; i < c3; i++ {
						h2.Remove(xx[i])
						if err := invariantImpl(h2); err != nil {
							chErr <- err
							return
						}
					}
					m3 := run(h2)
					var n3 int
					for k, v := range m3 {
						if m2[k] == v {
							n3++
						}
					}
					err := checkResults(float64(total-n3)/total, float64(c3)/float64(h2.Len()))
					if err != nil {
						chErr <- err
						return
					}
				}
			}
		}()
	}

	wg.Wait()
	select {
	case err := <-chErr:
		t.Fatal(err)
	default:
	}
}

func TestHash_Random(t *testing.T) {
	h := NewHash()
	if h.Random() != nil {
		t.Fatal("Random should return nil when h is empty")
	}

	for i := 0; i < 100; i++ {
		h.Add(i)
	}
	for i := 0; i < 10000; i++ {
		if v := h.Random(); v == nil || v.(int) < 0 || v.(int) >= 100 {
			t.Fatal("v == nil || v.(int) < 0 || v.(int) >= 100")
		}
	}
}
