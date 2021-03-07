package main

import (
	"context"
	_ "embed"

	"github.com/dim13/j1"
	"github.com/dim13/j1/console"
)

//go:embed j1e.bin
var eForth []byte

func main() {
	ctx, con := console.New(context.Background())
	vm := j1.New(con)
	vm.LoadBytes(eForth)
	vm.Run(ctx)
}
