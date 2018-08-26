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
	}
	return nil
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
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── value
//   └─────────────────────────────────────────────── 1
//
type Literal uint16

func newLiteral(v uint16) Literal { return Literal(v &^ uint16(1<<15)) }
func isLiteral(v uint16) bool     { return v&(1<<15) == 1<<15 }
func (v Literal) String() string  { return fmt.Sprintf("LIT %0.4X", uint16(v)) }
func (v Literal) value() uint16   { return uint16(v) }
func (v Literal) compile() uint16 { return v.value() | (1 << 15) }

// Jump instruction
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//   └──┴──┴───────────────────────────────────────── 0 0 0
//
type Jump uint16

func newJump(v uint16) Jump    { return Jump(v &^ uint16(7<<13)) }
func isJump(v uint16) bool     { return v&(7<<13) == 0<<13 }
func (v Jump) String() string  { return fmt.Sprintf("UBRANCH %0.4X", uint16(v<<1)) }
func (v Jump) value() uint16   { return uint16(v) }
func (v Jump) compile() uint16 { return v.value() }

// Conditional jump instruction
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//   └──┴──┴───────────────────────────────────────── 0 0 1
//
type Conditional uint16

func newConditional(v uint16) Conditional { return Conditional(v &^ uint16(7<<13)) }
func isConditional(v uint16) bool         { return v&(7<<13) == 1<<13 }
func (v Conditional) String() string      { return fmt.Sprintf("0BRANCH %0.4X", uint16(v<<1)) }
func (v Conditional) value() uint16       { return uint16(v) }
func (v Conditional) compile() uint16     { return v.value() | (1 << 13) }

// Call instruction
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  │  │  └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴── target
//   └──┴──┴───────────────────────────────────────── 0 1 0
//
type Call uint16

func newCall(v uint16) Call    { return Call(v &^ uint16(7<<13)) }
func isCall(v uint16) bool     { return v&(7<<13) == 2<<13 }
func (v Call) String() string  { return fmt.Sprintf("CALL %0.4X", uint16(v<<1)) }
func (v Call) value() uint16   { return uint16(v) }
func (v Call) compile() uint16 { return v.value() | (2 << 13) }

// ALU instruction
//
//  15 14 13 12 11 10  9  8  7  6  5  4  3  2  1  0
//   │  │  │  │  │  │  │  │  │  │  │  │  │  │  └──┴── dstack ± (see below)
//   │  │  │  │  │  │  │  │  │  │  │  │  └──┴──────── rstack ± (see below)
//   │  │  │  │  │  │  │  │  │  └──┴──┴────────────── modifier (see below)
//   │  │  │  │  │  │  │  │  └─────────────────────── RET
//   │  │  │  │  └──┴──┴──┴────────────────────────── opcode
//   │  │  │  └────────────────────────────────────── unused
//   └──┴──┴───────────────────────────────────────── 0 1 1
//
// Bits
//
//  654 modifier	32 rstack	10 dstack
//  001 = 1 T → N	01 +1		01 +1
//  010 = 2 T → R	10 -2		10 -2
//  011 = 3 N → [T]	11 -1		11 -1
//  100 = 4 N → io[T]
//
type ALU struct {
	Opcode Opcode
	Ret    bool
	Mod    Modifier
	Rdir   int8
	Ddir   int8
}

// expand 2 bit unsigned to 8 bit signed
var expand = map[uint16]int8{0: 0, 1: 1, 2: -2, 3: -1}

type Modifier uint8

const (
	ModTtoN     Modifier = 1
	ModTtoR     Modifier = 2
	ModNtoAtT   Modifier = 3
	ModNtoIoAtT Modifier = 4
)

func newALU(v uint16) ALU {
	return ALU{
		Opcode: Opcode((v >> 8) & 15),
		Ret:    v&(1<<7) != 0,
		Mod:    Modifier((v >> 4) & 7),
		Rdir:   expand[(v>>2)&3],
		Ddir:   expand[(v>>0)&3],
	}
}

func isALU(v uint16) bool { return v&(7<<13) == 3<<13 }

func (v ALU) value() uint16 {
	ret := uint16(v.Opcode) << 8
	if v.Ret {
		ret |= 1 << 7
	}
	ret |= uint16(v.Mod) << 4
	ret |= uint16(v.Rdir&3) << 2
	ret |= uint16(v.Ddir&3) << 0
	return ret
}

func (v ALU) compile() uint16 { return v.value() | (3 << 13) }

type Opcode uint16

const (
	OpT        Opcode = 0
	OpN        Opcode = 1
	OpTplusN   Opcode = 2
	OpTandN    Opcode = 3
	OpTorN     Opcode = 4
	OpTxorN    Opcode = 5
	OpNotT     Opcode = 6
	OpNeqT     Opcode = 7
	OpNleT     Opcode = 8
	OpNrshiftT Opcode = 9
	OpNlshiftT Opcode = 10
	OpR        Opcode = 11
	OpAtT      Opcode = 12
	OpIoAtT    Opcode = 13
	OpStatus   Opcode = 14
	OpNuLeT    Opcode = 15
)

var opcodeNames = []string{
	"T", "N", "T+N", "T∧N", "T∨N", "T⊻N", "¬T", "N=T",
	"N<T", "N≫T", "N≪T", "R", "[T]", "io[T]", "status", "Nu<T",
}

var modNames = []string{
	"", "T→N", "T→R", "N→[T]", "N→io[T]",
}

func (v ALU) String() string {
	s := "ALU " + opcodeNames[int(v.Opcode)]
	if v.Ret {
		s += " Ret"
	}
	s += modNames[int(v.Mod)]
	if v.Rdir != 0 {
		s += fmt.Sprintf(" r%+d", v.Rdir)
	}
	if v.Ddir != 0 {
		s += fmt.Sprintf(" d%+d", v.Ddir)
	}
	return s
}
