package main

import (
	"fmt"

	"dim13.org/j1"
)

func main() {
	body, err := j1.ReadBin("testdata/j1.bin")
	if err != nil {
		panic(err)
	}
	for i, v := range body {
		inst := j1.Decode(v)
		fmt.Printf("%0.4X %0.4X\t%s\n", 2*i, v, inst)
	}
}
