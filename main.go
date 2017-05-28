package main

import (
	"fmt"
)

func main() {
	vm := new(J1)
	if err := vm.ReadFile("testdata/j1.bin"); err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		vm.Eval()
	}
}

func dump() {
	body, err := ReadBin("testdata/j1.bin")
	if err != nil {
		panic(err)
	}
	for i, v := range body {
		inst := Decode(v)
		fmt.Printf("%0.4X %0.4X\t%s\n", 2*i, v, inst)
	}
}
