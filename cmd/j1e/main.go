package main

//go:generate file2go -in ../../testdata/j1e.bin

import (
	"dim13.org/j1"
	"dim13.org/j1/console"
)

func main() {
	con := console.New()
	defer con.Stop()
	vm := j1.New(con)
	vm.LoadBytes(J1eBin)
	vm.Run()
}
