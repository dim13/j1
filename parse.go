package j1

import "fmt"

// Decode instruction
func Decode(v uint16) Instruction {
	switch {
	case v&(1<<15) != 0:
		return newLit(v)
	case v&(7<<13) == 0:
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
	Value() uint16
	Compile() uint16
}

// Lit is a literal
type Lit uint16

func newLit(v uint16) Lit     { return Lit(v &^ uint16(1<<15)) }
func (v Lit) String() string  { return fmt.Sprintf("LIT %0.4X", uint16(v)) }
func (v Lit) Value() uint16   { return uint16(v) }
func (v Lit) Compile() uint16 { return v.Value() | (1 << 15) }

// Jump is an unconditional branch
type Jump uint16

func newJump(v uint16) Jump    { return Jump(v &^ uint16(7<<13)) }
func (v Jump) String() string  { return fmt.Sprintf("UBRANCH %0.4X", uint16(v<<1)) }
func (v Jump) Value() uint16   { return uint16(v) }
func (v Jump) Compile() uint16 { return v.Value() }

// Cond is a conditional branch
type Cond uint16

func newCond(v uint16) Cond    { return Cond(v &^ uint16(7<<13)) }
func (v Cond) String() string  { return fmt.Sprintf("0BRANCH %0.4X", uint16(v<<1)) }
func (v Cond) Value() uint16   { return uint16(v) }
func (v Cond) Compile() uint16 { return v.Value() | (1 << 13) }

// Call procedure
type Call uint16

func newCall(v uint16) Call    { return Call(v &^ uint16(7<<13)) }
func (v Call) String() string  { return fmt.Sprintf("CALL %0.4X", uint16(v<<1)) }
func (v Call) Value() uint16   { return uint16(v) }
func (v Call) Compile() uint16 { return v.Value() | (2 << 13) }

// ALU instruction
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

func (v ALU) Value() uint16 {
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
	ret |= uint16(v.Ddir & 3)
	return ret
}

func (v ALU) Compile() uint16 { return v.Value() | (3 << 13) }

func expand(v uint16) int8 {
	if v&2 != 0 {
		v |= 0xfc
	}
	return int8(v)
}

const (
	opT        = iota // 0
	opN               // 1
	opTplusN          // 2
	opTandN           // 3
	opTorN            // 4
	opTxorN           // 5
	opNotT            // 6
	opNeqT            // 7
	opNleT            // 8
	opNrshiftT        // 9
	opTminus1         // 10
	opR               // 11
	opAtT             // 12
	opNlshiftT        // 13
	opDepth           // 14
	opNuleT           // 15
)

var opcodeNames = []string{
	"T", "N", "T+N", "T&N", "T|N", "T^N", "~T", "N==T",
	"N<T", "N>>T", "T-1", "R", "[T]", "N<<T", "depth", "Nu<T",
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
