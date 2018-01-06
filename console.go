package j1

import (
	"bufio"
	"os"
)

type Console struct {
	r *bufio.Reader
	w *bufio.Writer
}

func NewConsole() *Console {
	return &Console{
		r: bufio.NewReader(os.Stdin),
		w: bufio.NewWriter(os.Stdout),
	}
}

func (c *Console) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

func (c *Console) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if err != nil {
		return 0, err
	}
	return n, c.w.Flush()
}
