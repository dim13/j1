package main

import (
	"context"

	"dim13.org/j1"
	"dim13.org/j1/console"
)

func main() {
	vm := j1.New()
	if err := vm.LoadFile("testdata/j1e.bin"); err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vm.Run(ctx, cancel, console.New(ctx))
}
