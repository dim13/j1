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
		fmt.Printf("%0.4X %0.4X\t%s\n", 2*i, v, op)
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
		return fmt.Sprintf("LIT %0.4X", v&0x7fff)
	case v&(7<<13) == 0:
		return fmt.Sprintf("UBRANCH %0.4X", v<<1)
	case v&(7<<13) == 1<<13:
		return fmt.Sprintf("0BRANCH %0.4X", v<<1)
	case v&(7<<13) == 1<<14:
		return fmt.Sprintf("CALL %0.4X", v<<1)
	case v&(7<<13) == 3<<13:
		op := (v >> 8) & 15
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
		switch expand((v >> 2) & 3) {
		case -1:
			s += " rstack-"
		case 1:
			s += " rstack+"
		}
		switch expand(v & 3) {
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
