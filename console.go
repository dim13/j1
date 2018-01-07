package j1

import (
	"io"
	"os"
)

type Console struct {
	r io.Reader
	w io.Writer
}

func NewConsole() *Console {
	return &Console{
		r: os.Stdin,
		w: os.Stdout,
	}
}

func (c *Console) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 && p[0] == 10 {
		p[0] = 13
	}
	return n, err
}

func (c *Console) Write(p []byte) (int, error) {
	return c.w.Write(p)
}
