package console

import (
	"fmt"
	"io"
	"os"
)

type Console struct {
	ich, och chan uint16
	done     chan struct{}
}

func New() *Console {
	c := &Console{
		ich:  make(chan uint16, 1),
		och:  make(chan uint16, 1),
		done: make(chan struct{}),
	}
	go c.read(os.Stdin)
	go c.write(os.Stdout)
	return c
}

func (c *Console) read(r io.Reader) {
	var v uint16
	defer close(c.ich)
	for {
		fmt.Fscanf(r, "%c", &v)
		select {
		case <-c.done:
			return
		case c.ich <- v:
		}
	}
}

func (c *Console) write(w io.Writer) {
	defer close(c.och)
	for {
		select {
		case <-c.done:
			return
		case v := <-c.och:
			fmt.Fprintf(w, "%c", v)
		}
	}
}

func (c *Console) Read() uint16   { return <-c.ich }
func (c *Console) Write(v uint16) { c.och <- v }
func (c *Console) Len() uint16    { return uint16(len(c.ich)) }
func (c *Console) Stop()          { close(c.done) }
