package main

//go:generate file2go -in ../../testdata/j1e.bin

import (
	"github.com/dim13/j1"
	"github.com/dim13/j1/console"
)

func main() {
	con := console.New()
	defer con.Stop()
	vm := j1.New(con)
	vm.LoadBytes(J1eBin)
	vm.Run()
}
