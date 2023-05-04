package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/dim13/j1"
)

func main() {
	body, err := ReadBin("testdata/j1e.bin")
	if err != nil {
		panic(err)
	}
	for i, v := range body {
		hi, lo := v>>8, v&0xff
		if hi < 0x20 || hi >= 0x7f {
			hi = 0x20
		}
		if lo < 0x20 || lo >= 0x7f {
			lo = 0x20
		}
		ins := j1.Decode(v)
		fmt.Printf("%0.4X %0.4X [%c%c]\t%s\n", 2*i, v, lo, hi, ins)
		if alu, ok := ins.(j1.ALU); ok && alu.RtoPC {
			fmt.Printf("\n")
		}
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
	if err := binary.Read(fd, binary.LittleEndian, &body); err != nil {
		return nil, err
	}
	return body, nil
}
