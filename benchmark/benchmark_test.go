package benchmark

import (
	"fmt"
	"testing"

	"github.com/edwingeng/doublejump"
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
