package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	fd, err := os.Open("testdata/j1.bin")
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	stat, err := fd.Stat()
	if err != nil {
		panic(err)
	}
	sz := stat.Size()
	body := make([]uint16, int(sz)/2)
	if err := binary.Read(fd, binary.BigEndian, &body); err != nil {
		panic(err)
	}
	for i, v := range body {
		op := Decode(v)
		fmt.Printf("%0.4x %0.16b %s\n", i, v, op)
	}
}

var opcodes = []string{
	"T",
	"N",
	"T+N",
	"TandN",
	"TorN",
	"TxorN",
	"~T",
	"N=T",
	"N<T",
	"NrshiftT",
	"T-1",
	"R",
	"[T]",
	"NlshiftT",
	"depth",
	"Nu<T",
}

func Decode(v uint16) string {
	switch {
	case v&(1<<15) == 1<<15:
		return fmt.Sprintf("literal %d", v&0x7fff)
	case v&(7<<13) == 0:
		return fmt.Sprintf("ubranch %0.4x", v&0x1fff)
	case v&(7<<13) == 1<<13:
		return fmt.Sprintf("0branch %0.4x", v&0x1fff)
	case v&(7<<13) == 1<<14:
		return fmt.Sprintf("call %0.4x", v&0x1fff)
	case v&(7<<13) == 3<<13:
		op := (v & 15 << 8) >> 8
		s := "ALU " + opcodes[op]
		if v&(1<<12) != 0 {
			s += " R→PC"
		}
		if v&(1<<7) != 0 {
			s += " T→N"
		}
		if v&(1<<6) != 0 {
			s += " T→R"
		}
		if v&(1<<5) != 0 {
			s += " N→[T]"
		}
		switch expand(v & (3 << 2) >> 2) {
		case -1:
			s += " rstack-"
		case 1:
			s += " rstack+"
		}
		switch expand(v & (3 << 0) >> 0) {
		case -1:
			s += " dstack-"
		case 1:
			s += " dstack+"
		}
		return s
	}
	return ""
}

func expand(v uint16) int8 {
	x := v >> 1
	for i := 7; i > 0; i-- {
		v |= x << uint(i)
	}
	return int8(v)
}