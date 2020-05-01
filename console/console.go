package console

import (
	"fmt"
	"io"
	"os"
)

// Console I/O
type Console struct {
	in, out chan uint16
	done    chan struct{}
}

// New console
func New() *Console {
	c := &Console{
		in:   make(chan uint16, 1),
		out:  make(chan uint16, 1),
		done: make(chan struct{}),
	}
	go c.write(os.Stdout)
	go c.read(os.Stdin)
	return c
}

func (c *Console) read(r io.Reader) {
	var v uint16
	defer close(c.in)
	for {
		_, err := fmt.Fscanf(r, "%c", &v)
		if err == io.EOF {
			return
		}
		select {
		case <-c.done:
			return
		case c.in <- v:
		}
	}
}

func (c *Console) write(w io.Writer) {
	defer close(c.out)
	for {
		select {
		case <-c.done:
			return
		case v := <-c.out:
			fmt.Fprintf(w, "%c", v)
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
func (c *Console) Stop() { close(c.done) }
