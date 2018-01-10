package main

//go:generate file2go -in ../../testdata/j1e.bin

import "dim13.org/j1"

func main() {
	vm := j1.New()
	vm.LoadBytes(J1eBin)
	vm.Eval()
}
