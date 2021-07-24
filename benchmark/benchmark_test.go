package benchmark

import (
	"fmt"
	"stathat.com/c/consistent"
	"strconv"
	"testing"

	"github.com/edwingeng/doublejump"
	"github.com/serialx/hashring"
)

var (
	g struct {
		Ret interface{}
	}
)

func BenchmarkDoubleJump(b *testing.B) {
	for i := 10; i <= 1000; i *= 10 {
		b.Run(fmt.Sprintf("%d-nodes", i), func(b *testing.B) {
			h := doublejump.NewHash()
			for j := 0; j < i; j++ {
				h.Add(fmt.Sprintf("node%d", j))
			}

			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				g.Ret = h.Get(uint64(j))
			}
		})
	}
}

func BenchmarkStathatConsistent(b *testing.B) {
	for i := 10; i <= 1000; i *= 10 {
		b.Run(fmt.Sprintf("%d-nodes", i), func(b *testing.B) {
			h := consistent.New()
			for j := 0; j < i; j++ {
				h.Add(fmt.Sprintf("node%d", j))
			}
			a := make([]string, b.N)
			for j := 0; j < b.N; j++ {
				a[j] = strconv.Itoa(j)
			}

			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				g.Ret, _ = h.Get(a[j])
			}
		})
	}
}

func BenchmarkSerialxHashring(b *testing.B) {
	for i := 10; i <= 1000; i *= 10 {
		b.Run(fmt.Sprintf("%d-nodes", i), func(b *testing.B) {
			nodes := make([]string, i)
			for j := 0; j < i; j++ {
				nodes[j] = fmt.Sprintf("node%d", j)
			}
			h := hashring.New(nodes)
			a := make([]string, b.N)
			for j := 0; j < b.N; j++ {
				a[j] = strconv.Itoa(j)
			}

			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				g.Ret, _ = h.GetNode(a[j])
			}
		})
	}
}
