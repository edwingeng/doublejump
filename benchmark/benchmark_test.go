package benchmark

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/edwingeng/doublejump"
	"github.com/serialx/hashring"
)

func BenchmarkDoubleJumpWithoutLock(b *testing.B) {
	for i := 10; i <= 1000; i *= 10 {
		b.Run(fmt.Sprintf("%d-nodes", i), func(b *testing.B) {
			h := doublejump.NewHash()
			for j := 0; j < i; j++ {
				h.Add(fmt.Sprintf("node%d", j))
			}

			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				h.Get(uint64(j))
			}
		})
	}
}

func BenchmarkDoubleJump(b *testing.B) {
	for i := 10; i <= 1000; i *= 10 {
		b.Run(fmt.Sprintf("%d-nodes", i), func(b *testing.B) {
			h := doublejump.NewHash()
			for j := 0; j < i; j++ {
				h.Add(fmt.Sprintf("node%d", j))
			}

			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				h.Get(uint64(j))
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
				h.GetNode(a[j])
			}
		})
	}
}
