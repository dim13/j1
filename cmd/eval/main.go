package main

import "dim13.org/j1"

func main() {
	vm := new(j1.J1)
	if err := vm.LoadFile("testdata/j1.bin"); err != nil {
		panic(err)
	}
	vm.Eval()
}
