package main

//go:generate file2go -in ../../testdata/j1e.bin

import (
	"context"

	"dim13.org/j1"
	"dim13.org/j1/console"
)

func main() {
	vm := j1.New()
	vm.LoadBytes(J1eBin)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vm.Run(ctx, cancel, console.New(ctx))
}
