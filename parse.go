package j1

import "fmt"

// Decode instruction from binary form
func Decode(v uint16) Instruction {
	switch {
	case isLiteral(v):
		return newLiteral(v)
	case isJump(v):
		return newJump(v)
	case isConditional(v):
		return newConditional(v)
	case isCall(v):
		return newCall(v)
	case isALU(v):
		return newALU(v)
	default:
		panic("invalid instruction")
	}
}

// Encode instruction to binary form
func Encode(i Instruction) uint16 {
	return i.compile()
}

// Instruction interface
type Instruction interface {
	value() uint16
	compile() uint16
}

// Literal value
//
//	15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//	 │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── value
//	 └─────────────────────────────────────────────── 1
type Literal uint16

func newLiteral(v uint16) Literal { return Literal(v &^ uint16(1<<15)) }
func isLiteral(v uint16) bool     { return v&(1<<15) == 1<<15 }
func (v Literal) String() string  { return fmt.Sprintf("LIT %0.4X", uint16(v)) }
func (v Literal) value() uint16   { return uint16(v) }
func (v Literal) compile() uint16 { return v.value() | (1 << 15) }

// Jump instruction
//
//	15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//	 │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//	 └──┴──┴───────────────────────────────────────── 0 0 0
type Jump uint16

func newJump(v uint16) Jump    { return Jump(v &^ uint16(7<<13)) }
func isJump(v uint16) bool     { return v&(7<<13) == 0 }
func (v Jump) String() string  { return fmt.Sprintf("JUMP %0.4X", uint16(v<<1)) }
func (v Jump) value() uint16   { return uint16(v) }
func (v Jump) compile() uint16 { return v.value() }

// Conditional jump instruction
//
//	15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//	 │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//	 └──┴──┴───────────────────────────────────────── 0 0 1
type Conditional uint16

func newConditional(v uint16) Conditional { return Conditional(v &^ uint16(7<<13)) }
func isConditional(v uint16) bool         { return v&(7<<13) == 1<<13 }
func (v Conditional) String() string      { return fmt.Sprintf("IF T=0 JUMP %0.4X", uint16(v<<1)) }
func (v Conditional) value() uint16       { return uint16(v) }
func (v Conditional) compile() uint16     { return v.value() | (1 << 13) }

// Call instruction
//
//	15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//	 │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//	 └──┴──┴───────────────────────────────────────── 0 1 0
type Call uint16

func newCall(v uint16) Call    { return Call(v &^ uint16(7<<13)) }
func isCall(v uint16) bool     { return v&(7<<13) == 2<<13 }
func (v Call) String() string  { return fmt.Sprintf("CALL %0.4X", uint16(v<<1)) }
func (v Call) value() uint16   { return uint16(v) }
func (v Call) compile() uint16 { return v.value() | (2 << 13) }

// ALU instruction
//
//	15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//	 │  │  │  │  │  │  │  │  │  │  │  │  │  │  └──┴── dstack ±
//	 │  │  │  │  │  │  │  │  │  │  │  │  └──┴──────── rstack ±
//	 │  │  │  │  │  │  │  │  │  │  │  └────────────── unused
//	 │  │  │  │  │  │  │  │  │  │  └───────────────── N → [T]
//	 │  │  │  │  │  │  │  │  │  └──────────────────── T → R
//	 │  │  │  │  │  │  │  │  └─────────────────────── T → N
//	 │  │  │  │  └──┴──┴──┴────────────────────────── Tʹ
//	 │  │  │  └────────────────────────────────────── R → PC
//	 └──┴──┴───────────────────────────────────────── 0 1 1
type ALU struct {
	Opcode Op
	RtoPC  bool
	TtoN   bool
	TtoR   bool
	NtoAtT bool
	Rdir   int8
	Ddir   int8
}

// expand 2 bit unsigned to 8 bit signed
var expand = []int8{0, 1, -2, -1}

func newALU(v uint16) ALU {
	return ALU{
		Opcode: Op(v>>8) & 15,
		RtoPC:  v&(1<<12) != 0,
		TtoN:   v&(1<<7) != 0,
		TtoR:   v&(1<<6) != 0,
		NtoAtT: v&(1<<5) != 0,
		Rdir:   expand[(v>>2)&3],
		Ddir:   expand[(v>>0)&3],
	}
}

func isALU(v uint16) bool { return v&(7<<13) == 3<<13 }

func (v ALU) value() uint16 {
	ret := uint16(v.Opcode) << 8
	if v.RtoPC {
		ret |= 1 << 12
	}
	if v.TtoN {
		ret |= 1 << 7
	}
	if v.TtoR {
		ret |= 1 << 6
	}
	if v.NtoAtT {
		ret |= 1 << 5
	}
	ret |= uint16(v.Rdir&3) << 2
	ret |= uint16(v.Ddir&3) << 0
	return ret
}

func (v ALU) compile() uint16 { return v.value() | (3 << 13) }

func (v ALU) String() string {
	s := fmt.Sprintf("T ← %v", v.Opcode)
	if v.RtoPC {
		s += " R→PC"
	}
	if v.TtoN {
		s += " T→N"
	}
	if v.TtoR {
		s += " T→R"
	}
	if v.NtoAtT {
		s += " N→[T]"
	}
	if v.Rdir != 0 {
		s += fmt.Sprintf(" r%+d", v.Rdir)
	}
	if v.Ddir != 0 {
		s += fmt.Sprintf(" d%+d", v.Ddir)
	}
	return s
}
