package j1

import (
	"fmt"
	"testing"
)

func cmp(t *testing.T, got, want Core) {
	t.Helper()
	if got.pc != want.pc {
		t.Errorf("pc: got %0.4X, want %0.4X", got.pc, want.pc)
	}
	if got.st0 != want.st0 {
		t.Errorf("st0: got %0.4X, want %0.4X", got.st0, want.st0)
	}
	if got.d.sp != want.d.sp {
		t.Errorf("dsp: got %0.4X, want %0.4X", got.d.sp, want.d.sp)
	}
	if got.r.sp != want.r.sp {
		t.Errorf("rsp: got %0.4X, want %0.4X", got.r.sp, want.r.sp)
	}
	if got.d.data != want.d.data {
		t.Errorf("d:stack: got %0.4X, want %0.4X", got.d.data, want.d.data)
	}
	if got.r.data != want.r.data {
		t.Errorf("rstack: got %0.4X, want %0.4X", got.r.data, want.r.data)
	}
}

type mocConsole struct{}

func (m *mocConsole) Read() uint16 { return 0 }
func (m *mocConsole) Write(uint16) {}
func (m *mocConsole) Len() uint16  { return 0 }

func TestEval(t *testing.T) {
	testCases := []struct {
		ins []Instruction
		end Core
	}{
		{
			ins: []Instruction{Jump(0xff)},
			end: Core{pc: 0xff},
		},
		{
			ins: []Instruction{Lit(1), Cond(0xff)},
			end: Core{pc: 2},
		},
		{
			ins: []Instruction{Lit(0), Cond(0xff)},
			end: Core{pc: 0xff},
		},
		{
			ins: []Instruction{Call(0xff)},
			end: Core{pc: 0xff, r: stack{data: [0x20]uint16{0x00, 0x02}, sp: 1}},
		},
		{
			ins: []Instruction{Lit(0xff)},
			end: Core{pc: 1, st0: 0xff, d: stack{sp: 1}},
		},
		{
			ins: []Instruction{Lit(0xff), Lit(0xfe)},
			end: Core{pc: 2, st0: 0xfe, d: stack{data: [0x20]uint16{0x00, 0x00, 0xff}, sp: 2}},
		},
		{ // dup
			ins: []Instruction{Lit(0xff), ALU{Opcode: opT, TtoN: true, Ddir: 1}},
			end: Core{pc: 2, st0: 0xff, d: stack{data: [0x20]uint16{0x00, 0x00, 0xff}, sp: 2}},
		},
		{ // over
			ins: []Instruction{Lit(0xaa), Lit(0xbb), ALU{Opcode: opN, TtoN: true, Ddir: 1}},
			end: Core{pc: 3, st0: 0xaa, d: stack{data: [0x20]uint16{0x00, 0x00, 0xaa, 0xbb}, sp: 3}},
		},
		{ // invert
			ins: []Instruction{Lit(0x00ff), ALU{Opcode: opNotT}},
			end: Core{pc: 2, st0: 0xff00, d: stack{sp: 1}},
		},
		{ // +
			ins: []Instruction{Lit(1), Lit(2), ALU{Opcode: opTplusN, Ddir: -1}},
			end: Core{pc: 3, st0: 3, d: stack{data: [0x20]uint16{0, 0, 1}, sp: 1}},
		},
		{ // swap
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opN, TtoN: true}},
			end: Core{pc: 3, st0: 2, d: stack{data: [0x20]uint16{0, 0, 3}, sp: 2}},
		},
		{ // nip
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opT, Ddir: -1}},
			end: Core{pc: 3, st0: 3, d: stack{data: [0x20]uint16{0, 0, 2}, sp: 1}},
		},
		{ // drop
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opN, Ddir: -1}},
			end: Core{pc: 3, st0: 2, d: stack{data: [0x20]uint16{0, 0, 2}, sp: 1}},
		},
		{ // ;
			ins: []Instruction{Call(10), Call(20), ALU{Opcode: opT, RtoPC: true, Rdir: -1}},
			end: Core{pc: 11, r: stack{data: [0x20]uint16{0, 2, 22}, sp: 1}},
		},
		{ // >r
			ins: []Instruction{Lit(10), ALU{Opcode: opN, TtoR: true, Ddir: -1, Rdir: 1}},
			end: Core{pc: 2, r: stack{data: [0x20]uint16{0, 10}, sp: 1}},
		},
		{ // r>
			ins: []Instruction{Lit(10), Call(20), ALU{Opcode: opR, TtoN: true, TtoR: true, Ddir: 1, Rdir: -1}},
			end: Core{pc: 21, st0: 4, d: stack{data: [0x20]uint16{0, 0, 10}, sp: 2}, r: stack{data: [0x20]uint16{10, 4}}},
		},
		{ // r@
			ins: []Instruction{Lit(10), ALU{Opcode: opR, TtoN: true, TtoR: true, Ddir: 1}},
			end: Core{pc: 2, d: stack{data: [0x20]uint16{0, 0, 10}, sp: 2}, r: stack{data: [0x20]uint16{10}}},
		},
		{ // @
			ins: []Instruction{ALU{Opcode: opAtT}},
			end: Core{pc: 1},
		},
		{ // !
			ins: []Instruction{Lit(1), Lit(0), ALU{Opcode: opN, NtoAtT: true, Ddir: -1}},
			end: Core{pc: 3, st0: 1, d: stack{data: [0x20]uint16{0, 0, 1}, sp: 1}, memory: [memSize]uint16{1}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			state := New(&mocConsole{})
			for _, ins := range tc.ins {
				state.Eval(ins)
			}
			cmp(t, *state, tc.end)
		})
	}
}

func TestNextST0(t *testing.T) {
	testCases := []struct {
		ins   ALU
		st0   uint16
		state Core
	}{
		{ins: ALU{Opcode: opT}, st0: 0xff, state: Core{st0: 0xff}},
		{ins: ALU{Opcode: opN}, st0: 0xbb, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xbb}, sp: 2}}},
		{ins: ALU{Opcode: opTplusN}, st0: 0x01ba, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xbb}, sp: 2}}},
		{ins: ALU{Opcode: opTandN}, st0: 0xbb, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xbb}, sp: 2}}},
		{ins: ALU{Opcode: opTorN}, st0: 0xff, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xbb}, sp: 2}}},
		{ins: ALU{Opcode: opTxorN}, st0: 0x44, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xbb}, sp: 2}}},
		{ins: ALU{Opcode: opNotT}, st0: 0xff55, state: Core{st0: 0xaa}},
		{ins: ALU{Opcode: opNeqT}, st0: 0x00, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xbb}, sp: 2}}},
		{ins: ALU{Opcode: opNeqT}, st0: 0xffff, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xff}, sp: 2}}},
		{ins: ALU{Opcode: opNleT}, st0: 0xffff, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xbb}, sp: 2}}},
		{ins: ALU{Opcode: opNleT}, st0: 0x00, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xff}, sp: 2}}},
		{ins: ALU{Opcode: opNrshiftT}, st0: 0x3f, state: Core{st0: 0x02, d: stack{data: [0x20]uint16{0, 0xaa, 0xff}, sp: 2}}},
		{ins: ALU{Opcode: opTminus1}, st0: 0x54, state: Core{st0: 0x55}},
		{ins: ALU{Opcode: opR}, st0: 0x5, state: Core{r: stack{data: [0x20]uint16{0, 0x05}, sp: 1}}},
		{ins: ALU{Opcode: opAtT}, st0: 0x5, state: Core{st0: 0x02, memory: [memSize]uint16{0, 5, 10}}},
		{ins: ALU{Opcode: opNlshiftT}, st0: 0x3fc, state: Core{st0: 0x02, d: stack{data: [0x20]uint16{0, 0xaa, 0xff}, sp: 2}}},
		{ins: ALU{Opcode: opDepth}, st0: 0x305, state: Core{r: stack{sp: 3}, d: stack{sp: 5}}},
		{ins: ALU{Opcode: opNuleT}, st0: 0xffff, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xbb}, sp: 2}}},
		{ins: ALU{Opcode: opNuleT}, st0: 0x00, state: Core{st0: 0xff, d: stack{data: [0x20]uint16{0, 0xaa, 0xff}, sp: 2}}},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			state := &tc.state
			st0 := state.newST0(tc.ins.Opcode)
			if st0 != tc.st0 {
				t.Errorf("got %x, want %x", st0, tc.st0)
			}
		})
	}
}

func TestLoadBytes(t *testing.T) {
	data := []byte{1, 2, 4, 8}
	j1 := New(&mocConsole{})
	if err := j1.LoadBytes(data); err != nil {
		t.Fatal(err)
	}
	expect := [memSize]uint16{0x0201, 0x0804}
	if j1.memory != expect {
		t.Errorf("got %v, want %v", j1.memory[:2], expect)
	}
}

func TestReset(t *testing.T) {
	j1 := &Core{pc: 100, d: stack{sp: 2}, r: stack{sp: 3}, st0: 5}
	j1.Reset()
	if j1.pc != 0 || j1.d.sp != 0 || j1.r.sp != 0 || j1.st0 != 0 {
		t.Errorf("got %v", j1)
	}
}
