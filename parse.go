package j1

import "fmt"

// Decode instruction
func Decode(v uint16) Instruction {
	switch {
	case v&(1<<15) == 1<<15:
		return newLit(v)
	case v&(7<<13) == 0<<13:
		return newJump(v)
	case v&(7<<13) == 1<<13:
		return newCond(v)
	case v&(7<<13) == 2<<13:
		return newCall(v)
	case v&(7<<13) == 3<<13:
		return newALU(v)
	}
	return nil
}

// Instruction interface
type Instruction interface {
	value() uint16
	compile() uint16
}

// Lit is a literal
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── value
//   └─────────────────────────────────────────────── 1
//
type Lit uint16

func newLit(v uint16) Lit     { return Lit(v &^ uint16(1<<15)) }
func (v Lit) String() string  { return fmt.Sprintf("LIT %0.4X", uint16(v)) }
func (v Lit) value() uint16   { return uint16(v) }
func (v Lit) compile() uint16 { return v.value() | (1 << 15) }

// Jump is an unconditional branch
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//   └──┴──┴───────────────────────────────────────── 0 0 0
//
type Jump uint16

func newJump(v uint16) Jump    { return Jump(v &^ uint16(7<<13)) }
func (v Jump) String() string  { return fmt.Sprintf("UBRANCH %0.4X", uint16(v<<1)) }
func (v Jump) value() uint16   { return uint16(v) }
func (v Jump) compile() uint16 { return v.value() }

// Cond is a conditional branch
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//   └──┴──┴───────────────────────────────────────── 0 0 1
//
type Cond uint16

func newCond(v uint16) Cond    { return Cond(v &^ uint16(7<<13)) }
func (v Cond) String() string  { return fmt.Sprintf("0BRANCH %0.4X", uint16(v<<1)) }
func (v Cond) value() uint16   { return uint16(v) }
func (v Cond) compile() uint16 { return v.value() | (1 << 13) }

// Call procedure
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//   └──┴──┴───────────────────────────────────────── 0 1 0
//
type Call uint16

func newCall(v uint16) Call    { return Call(v &^ uint16(7<<13)) }
func (v Call) String() string  { return fmt.Sprintf("CALL %0.4X", uint16(v<<1)) }
func (v Call) value() uint16   { return uint16(v) }
func (v Call) compile() uint16 { return v.value() | (2 << 13) }

// ALU instruction
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  │  │  │  │  │  │  │  │  │  │  │  │  │  └──┴── dstack ±
//   │  │  │  │  │  │  │  │  │  │  │  │  └──┴──────── rstack ±
//   │  │  │  │  │  │  │  │  │  │  │  └────────────── unused
//   │  │  │  │  │  │  │  │  │  │  └───────────────── N → [T]
//   │  │  │  │  │  │  │  │  │  └──────────────────── T → R
//   │  │  │  │  │  │  │  │  └─────────────────────── T → N
//   │  │  │  │  └──┴──┴──┴────────────────────────── Tʹ
//   │  │  │  └────────────────────────────────────── R → PC
//   └──┴──┴───────────────────────────────────────── 0 1 1
//
type ALU struct {
	Opcode uint16
	RtoPC  bool
	TtoN   bool
	TtoR   bool
	NtoAtT bool
	Rdir   int8
	Ddir   int8
}

// expand 2 bit unsigned to 8 bit signed
var expand = map[uint16]int8{0: 0, 1: 1, 2: -2, 3: -1}

func newALU(v uint16) ALU {
	return ALU{
		Opcode: (v >> 8) & 15,
		RtoPC:  v&(1<<12) != 0,
		TtoN:   v&(1<<7) != 0,
		TtoR:   v&(1<<6) != 0,
		NtoAtT: v&(1<<5) != 0,
		Rdir:   expand[(v>>2)&3],
		Ddir:   expand[(v>>0)&3],
	}
}

func (v ALU) value() uint16 {
	ret := v.Opcode << 8
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

const (
	opT        = 0x0
	opN        = 0x1
	opTplusN   = 0x2
	opTandN    = 0x3
	opTorN     = 0x4
	opTxorN    = 0x5
	opNotT     = 0x6
	opNeqT     = 0x7
	opNleT     = 0x8
	opNrshiftT = 0x9
	opTminus1  = 0xa
	opR        = 0xb
	opAtT      = 0xc
	opNlshiftT = 0xd
	opDepth    = 0xe
	opNuleT    = 0xf
)

var opcodeNames = []string{
	"T", "N", "T+N", "T∧N", "T∨N", "T⊻N", "¬T", "N=T",
	"N<T", "N≫T", "T-1", "R", "[T]", "N≪T", "depth", "Nu<T",
}

func (v ALU) String() string {
	s := "ALU " + opcodeNames[v.Opcode]
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
