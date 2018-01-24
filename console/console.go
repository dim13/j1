package console

import (
	"fmt"
	"io"
	"os"
)

type Console struct {
	r        io.Reader
	w        io.Writer
	ich, och chan uint16
	done     chan struct{}
}

func New() *Console {
	c := &Console{
		r:    os.Stdin,
		w:    os.Stdout,
		ich:  make(chan uint16, 1),
		och:  make(chan uint16, 1),
		done: make(chan struct{}),
	}
	go c.read()
	go c.write()
	return c
}

func (c *Console) read() {
	var v uint16
	for {
		fmt.Fscanf(c.r, "%c", &v)
		select {
		case <-c.done:
			return
		case c.ich <- v:
		}
	}
}

func (c *Console) write() {
	for {
		select {
		case <-c.done:
			return
		case v := <-c.och:
			fmt.Fprintf(c.w, "%c", v)
		}
	}
}

func (c *Console) Read() uint16   { return <-c.ich }
func (c *Console) Write(v uint16) { c.och <- v }
func (c *Console) Len() uint16    { return uint16(len(c.ich)) }
func (c *Console) Stop()          { close(c.done) }
