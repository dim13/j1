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
	"T&N",
	"T|N",
	"T^N",
	"~T",
	"N==T",
	"N<T",
	"N>>T",
	"T-1",
	"rT",
	"[T]",
	"N<<T",
	"dsp",
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
			s += " r-1"
		case -2:
			s += " r-2" // ???
		case 1:
			s += " r+1"
		}
		switch expand(v & 3) {
		case -1:
			s += " d-1"
		case 1:
			s += " d+1"
		}
		return s
	}
	return ""
}

func expand(v uint16) int8 {
	switch v {
	case 0: // 00 → 00000000
		return 0
	case 1: // 01 → 00000001
		return 1
	case 2: // 10 → 11111110
		return -2
	case 3: // 11 → 11111111
		return -1
	}
	return 0
}
