package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"dim13.org/j1"
)

func main() {
	body, err := ReadBin("testdata/j1e.bin")
	if err != nil {
		panic(err)
	}
	for i, v := range body {
		fmt.Printf("%0.4X %0.4X\t%s\n", 2*i, v, j1.Decode(v))
	}
}

// ReadBin file
func ReadBin(fname string) ([]uint16, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	stat, err := fd.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	body := make([]uint16, int(size)/2)
	if err := binary.Read(fd, binary.BigEndian, &body); err != nil {
		return nil, err
	}
	return body, nil
}
