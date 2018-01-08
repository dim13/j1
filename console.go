package j1

import (
	"io"
	"os"
)

type Console struct {
	io.Reader
	io.Writer
}

func NewConsole() *Console {
	return &Console{Reader: os.Stdin, Writer: os.Stdout}
}
