package main

import "fmt"

func editMap(m map[int]int) {
	m[3] = 8
	m[4] = 9
}

func editSlice(s []int) {
	s = append(s, 5)
	s = append(s, 9)
}

func main() {
	m := make(map[int]int, 0)
	m[1] = 5
	m[2] = 3
	fmt.Printf("%v\n", m)
	editMap(m)
	fmt.Printf("%v\n", m)

	s := make([]int, 2)
	s[0] = 2
	s[1] = 5
	fmt.Printf("%v\n", s)
	editSlice(s)
	fmt.Printf("%v\n", s)
}
