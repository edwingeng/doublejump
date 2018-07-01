# Overview
This is a variation of the Google's jump consistent hash. It overcomes the shortcoming of not being able to remove nodes of the original implementation at the price of not being so consistent under some circumstances.

# Benchmark
```
BenchmarkDoubleJump/10-nodes            20000000            73.3 ns/op
BenchmarkDoubleJump/100-nodes           20000000            85.9 ns/op
BenchmarkDoubleJump/1000-nodes          20000000            97.5 ns/op

BenchmarkStathatConsistent/10-nodes      5000000           293 ns/op
BenchmarkStathatConsistent/100-nodes     5000000           327 ns/op
BenchmarkStathatConsistent/1000-nodes    3000000           435 ns/op

BenchmarkSerialxHashring/10-nodes        5000000           283 ns/op
BenchmarkSerialxHashring/100-nodes       5000000           341 ns/op
BenchmarkSerialxHashring/1000-nodes      3000000           426 ns/op
```

# Example
```
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
```
