package j1

import (
	"encoding/binary"
	"fmt"
	"os"
)

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

func Decode(v uint16) Instruction {
	switch {
	case v&(1<<15) != 0:
		return newLit(v)
	case v&(7<<13) == 0:
		return newJump(v)
	case v&(7<<13) == 1<<13:
		return newCond(v)
	case v&(7<<13) == 1<<14:
		return newCall(v)
	case v&(7<<13) == 3<<13:
		return newALU(v)
	}
	return nil
}

type Instruction interface {
	isInstruction()
}

type Lit uint16

func newLit(v uint16) Lit    { return Lit(v &^ uint16(1<<15)) }
func (v Lit) String() string { return fmt.Sprintf("LIT %0.4X", uint16(v)) }
func (v Lit) isInstruction() {}

type Jump uint16

func newJump(v uint16) Jump   { return Jump(v &^ uint16(7<<13)) }
func (v Jump) String() string { return fmt.Sprintf("UBRANCH %0.4X", uint16(v<<1)) }
func (v Jump) isInstruction() {}

type Cond uint16

func newCond(v uint16) Cond   { return Cond(v &^ uint16(7<<13)) }
func (v Cond) String() string { return fmt.Sprintf("0BRANCH %0.4X", uint16(v<<1)) }
func (v Cond) isInstruction() {}

type Call uint16

func newCall(v uint16) Call   { return Call(v &^ uint16(7<<13)) }
func (v Call) String() string { return fmt.Sprintf("CALL %0.4X", uint16(v<<1)) }
func (v Call) isInstruction() {}

type ALU struct {
	Opcode uint16
	RtoPC  bool
	TtoN   bool
	TtoR   bool
	NtoAtT bool
	Rdir   int8
	Ddir   int8
}

func newALU(v uint16) ALU {
	return ALU{
		Opcode: (v >> 8) & 15,
		RtoPC:  v&(1<<12) != 0,
		TtoN:   v&(1<<7) != 0,
		TtoR:   v&(1<<6) != 0,
		NtoAtT: v&(1<<5) != 0,
		Rdir:   expand((v >> 2) & 3),
		Ddir:   expand(v & 3),
	}
}

func (v ALU) isInstruction() {}

func expand(v uint16) int8 {
	if v&2 != 0 {
		v |= 0xfc
	}
	return int8(v)
}

var opcodes = []string{
	"T", "N", "T+N", "T&N", "T|N", "T^N", "~T", "N==T",
	"N<T", "N>>T", "T-1", "R", "[T]", "N<<T", "depth", "Nu<T",
}

func (a ALU) String() string {
	s := "ALU " + opcodes[a.Opcode]
	if a.RtoPC {
		s += " R→PC"
	}
	if a.TtoN {
		s += " T→N"
	}
	if a.TtoR {
		s += " T→R"
	}
	if a.NtoAtT {
		s += " N→[T]"
	}
	if a.Rdir != 0 {
		s += fmt.Sprintf(" r%+d", a.Rdir)
	}
	if a.Ddir != 0 {
		s += fmt.Sprintf(" d%+d", a.Ddir)
	}
	return s
}
