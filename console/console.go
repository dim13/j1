package console

import (
	"context"
	"fmt"
	"io"
	"os"
)

type Console struct {
	r        io.Reader
	w        io.Writer
	ich, och chan uint16
}

func New(ctx context.Context) *Console {
	c := &Console{
		r:   os.Stdin,
		w:   os.Stdout,
		ich: make(chan uint16, 1),
		och: make(chan uint16, 1),
	}
	go c.read(ctx)
	go c.write(ctx)
	return c
}

func (c *Console) read(ctx context.Context) {
	var v uint16
	for {
		fmt.Fscanf(c.r, "%c", &v)
		select {
		case <-ctx.Done():
			return
		case c.ich <- v:
		}
	}
}

func (c *Console) write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case v := <-c.och:
			fmt.Fprintf(c.w, "%c", v)
		}
	}
}

func (c *Console) Read() uint16   { return <-c.ich }
func (c *Console) Write(v uint16) { c.och <- v }
func (c *Console) Len() uint16    { return uint16(len(c.ich)) }
