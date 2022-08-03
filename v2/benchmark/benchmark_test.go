package benchmark

import (
	"fmt"
	"testing"

	"github.com/edwingeng/doublejump/v2"
)

var (
	g struct {
		Ret string
	}
)

func BenchmarkDoublejump(b *testing.B) {
	for i := 10; i <= 1000; i *= 10 {
		b.Run(fmt.Sprintf("%d-nodes", i), func(b *testing.B) {
			h := doublejump.NewHash[string]()
			for j := 0; j < i; j++ {
				h.Add(fmt.Sprintf("node%d", j))
			}

			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				g.Ret, _ = h.Get(uint64(j))
			}
		})
	}
}
