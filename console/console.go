package console

import "fmt"

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
	go c.write()
	go c.read()
	return c
}

func (c *Console) read() {
	defer close(c.in)
	for {
		var v uint16
		fmt.Scanf("%c", &v)
		select {
		case <-c.done:
			return
		case c.in <- v:
		}
	}
}

func (c *Console) write() {
	defer close(c.out)
	for {
		select {
		case <-c.done:
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
func (c *Console) Stop() { close(c.done) }
