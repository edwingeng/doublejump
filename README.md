# Overview
This is a variation of the Google's jump consistent hash. It overcomes the shortcoming of not being able to remove nodes of the original implementation at the price of not being so consistent under some circumstances.

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
