# Overview
Doublejump is a revamped [Google's jump](https://arxiv.org/pdf/1406.2294.pdf) consistent hash. It overcomes the shortcoming of the original design - being unable to remove nodes. Here is [how it works](https://docs.google.com/presentation/d/e/2PACX-1vTHyFGUJ5CBYxZTzToc_VKxP_Za85AeZqQMNGLXFLP1tX0f9IF_z3ys9-pyKf-Jj3iWpm7dUDDaoFyb/pub?start=false&loop=false&delayms=3000).

# Benchmark
```
BenchmarkDoubleJump/10-nodes                    48824548                22.3 ns/op
BenchmarkDoubleJump/100-nodes                   33921781                34.9 ns/op
BenchmarkDoubleJump/1000-nodes                  25635931                46.3 ns/op

BenchmarkStathatConsistent/10-nodes              4961104               245.6 ns/op
BenchmarkStathatConsistent/100-nodes             4507544               284.1 ns/op
BenchmarkStathatConsistent/1000-nodes            3412558               358.2 ns/op

BenchmarkSerialxHashring/10-nodes                2816680               445.7 ns/op
BenchmarkSerialxHashring/100-nodes               2535745               482.1 ns/op
BenchmarkSerialxHashring/1000-nodes              2243271               549.6 ns/op
```

# Examples

### V1
```go
// If golang version <= 1.17
import "github.com/edwingeng/doublejump"

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
```

### V2
```go
// If golang version >= 1.18
import "github.com/edwingeng/doublejump/v2"

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
```

# Acknowledgements
The implementation of the original algorithm is credited to [dgryski](https://github.com/dgryski/go-jump).
