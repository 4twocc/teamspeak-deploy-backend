package main

import (
	"fmt"
	"sort"
	"test/utils"
	"unicode"
)

type Map map[string][]int

var (
	a int32   = 1
	b string  = "1"
	c bool    = true
	d float64 = 1.0
	e []int32

	f string = "这是个一包含汉字和英文的字符串, This is an apple"

	count int = 0

	m [5]int = [5]int{1, 2, 3, 4, 5}
	x []int  = []int{1, 2, 3, 4, 5}
)

var mm = [...]int{3, 7, 8, 9, 1}

func init() {
	fmt.Println("init")
	fmt.Printf("切片X的容量是: %d, 切片X的长度是: %d\n", (cap(x)), len(x))
}
func main() {
	fmt.Println("Hello World")
	fmt.Println(utils.GetEnv("HOME", "/home/"))

	fmt.Printf("a is %T\n", a)
	fmt.Printf("b is %T\n", b)
	fmt.Printf("c is %T\n", c)
	fmt.Printf("d is %T\n", d)
	fmt.Printf("e is %T\n", e)

	for _, i := range f {
		// i is rune
		if unicode.Is(unicode.Han, i) {
			count++
		}
	}
	fmt.Printf("中文的数量是: %d\n", count)

	temp := []int{1, 1, 2, 2, 3, 4, 4, 5, 5}

	n := temp[0]

	for _, i := range temp {
		n ^= temp[i]
	}
	fmt.Println(n)
	fmt.Println(sum())
	fmt.Println(s_index())

	m := make(Map)
	s := []int{1, 2}
	s = append(s, 3)
	fmt.Printf("%+v\n", s)
	m["q1mi"] = s
	fmt.Printf("===%+v\n", s[:1])
	fmt.Printf("===%+v\n", s[2:])
	s = append(s[:1], s[2:]...)
	fmt.Printf("%+v\n", s)
}

// func print_n_multi() {
// 	for i := 1; i < 10; i++ {
// 		for j := i; j < 10; j++ {
// 			fmt.Printf("%d * %d = %d\n", i, j, i*j)
// 		}
// 	}
// }

func sum() int {
	sum := 0
	for _, i := range m {
		sum += i
	}
	return sum
}

func s_index() []int {
	sort.Ints(mm[:])
	return mm[:]
}
