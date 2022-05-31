package doublejump

import "fmt"

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
