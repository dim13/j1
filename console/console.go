package console

import (
	"context"
	"fmt"
	"io"
)

// Console I/O
type Console struct {
	in, out chan uint16
	ctx     context.Context
	cancel  func()
}

// New console
func New(ctx context.Context) (context.Context, *Console) {
	c := &Console{
		in:  make(chan uint16, 1),
		out: make(chan uint16, 1),
	}
	ctx, c.cancel = context.WithCancel(ctx)
	go c.write(ctx)
	go c.read(ctx)
	return ctx, c
}

func (c *Console) read(ctx context.Context) {
	defer close(c.in)
	for {
		var v uint16
		_, err := fmt.Scanf("%c", &v)
		if err == io.EOF {
			c.Stop()
		}
		select {
		case <-ctx.Done():
			return
		case c.in <- v:
		}
	}
}

func (c *Console) write(ctx context.Context) {
	defer close(c.out)
	for {
		select {
		case <-ctx.Done():
			return
		case v := <-c.out:
			fmt.Printf("%c", v)
		}
	}
}

// Read from console
func (c *Console) Read() uint16 { return <-c.in }

// Write to console
func (c *Console) Write(v uint16) { c.out <- v }

// Len of input buffer
func (c *Console) Len() uint16 { return uint16(len(c.in)) }

// Stop console
func (c *Console) Stop() { c.cancel() }
