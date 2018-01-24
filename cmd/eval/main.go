package main

import (
	"dim13.org/j1"
	"dim13.org/j1/console"
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
