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
	n, err := c.r.Read(p)
	if n > 0 && p[0] == 10 {
		p[0] = 13
	}
	return n, err
}

func (c *Console) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if err != nil {
		return 0, err
	}
	return n, c.w.Flush()
}
