package console

import (
	"context"
	"fmt"
	"io"
)

// Console I/O
type Console struct {
	input  chan uint16
	cancel func()
}

// New console
func New(ctx context.Context) (context.Context, *Console) {
	ctx, cancel := context.WithCancel(ctx)
	c := &Console{
		input:  make(chan uint16, 1),
		cancel: cancel,
	}
	go func() {
		var v uint16
		defer close(c.input)
		for {
			_, err := fmt.Scanf("%c", &v)
			if err == io.EOF {
				cancel()
			}
			select {
			case <-ctx.Done():
				return
			case c.input <- v:
			}
		}
	}()
	return ctx, c
}

// Read from console
func (c *Console) Read() uint16 {
	return <-c.input
}

// Write to console
func (c *Console) Write(v uint16) {
	fmt.Printf("%c", v)
}

// Len of input buffer
func (c *Console) Len() uint16 {
	if len(c.input) > 0 {
		return 1
	}
	return 0
}

// Stop console
func (c *Console) Stop() {
	c.cancel()
}
