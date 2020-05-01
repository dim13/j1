package main

import (
	"github.com/dim13/j1"
	"github.com/dim13/j1/console"
)

func main() {
	con := console.New()
	defer con.Stop()
	vm := j1.New(con)
	if err := vm.LoadFile("testdata/j1e.bin"); err != nil {
		panic(err)
	}
	vm.Run()
}
