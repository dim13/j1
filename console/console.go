package console

import (
	"fmt"
	"io"
)

// Console I/O
type Console struct {
	ich, och chan uint16
	done     chan struct{}
}

// New console
func New(w io.Writer, r io.Reader) *Console {
	c := &Console{
		ich:  make(chan uint16, 1),
		och:  make(chan uint16, 1),
		done: make(chan struct{}),
	}
	go c.write(w)
	go c.read(r)
	return c
}

func (c *Console) read(r io.Reader) {
	var v uint16
	defer close(c.ich)
	for {
		_, err := fmt.Fscanf(r, "%c", &v)
		if err == io.EOF {
			return
		}
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

// Read from console
func (c *Console) Read() uint16 { return <-c.ich }

// Write to console
func (c *Console) Write(v uint16) { c.och <- v }

// Len of input buffer
func (c *Console) Len() uint16 { return uint16(len(c.ich)) }

// Stop console
func (c *Console) Stop() { close(c.done) }
