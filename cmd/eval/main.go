package main

import "dim13.org/j1"

func main() {
	vm := j1.New()
	if err := vm.LoadFile("testdata/j1e.bin"); err != nil {
		panic(err)
	}
	vm.Run()
}
