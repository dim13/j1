package main

import (
	"fmt"
)

func main() {
	body, err := ReadBin("testdata/j1.bin")
	if err != nil {
		panic(err)
	}
	for i, v := range body {
		inst := Decode(v)
		fmt.Printf("%0.4X %0.4X\t%s\n", 2*i, v, inst)
	}
}
