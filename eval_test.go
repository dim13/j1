package j1

import (
	"fmt"
	"testing"
)

func cmp(t *testing.T, got, want J1) {
	t.Helper()
	if got.pc != want.pc {
		t.Errorf("pc: got %0.4X, want %0.4X", got.pc, want.pc)
	}
	if got.st0 != want.st0 {
		t.Errorf("st0: got %0.4X, want %0.4X", got.st0, want.st0)
	}
	if got.dsp != want.dsp {
		t.Errorf("dsp: got %0.4X, want %0.4X", got.dsp, want.dsp)
	}
	if got.rsp != want.rsp {
		t.Errorf("rsp: got %0.4X, want %0.4X", got.rsp, want.rsp)
	}
	if got.dstack != want.dstack {
		t.Errorf("dstack: got %0.4X, want %0.4X", got.dstack, want.dstack)
	}
	if got.rstack != want.rstack {
		t.Errorf("rstack: got %0.4X, want %0.4X", got.rstack, want.rstack)
	}
}

func TestEval(t *testing.T) {
	testCases := []struct {
		ins []Instruction
		end J1
	}{
		{
			ins: []Instruction{Jump(0xff)},
			end: J1{pc: 0xff},
		},
		{
			ins: []Instruction{Lit(1), Cond(0xff)},
			end: J1{pc: 2},
		},
		{
			ins: []Instruction{Lit(0), Cond(0xff)},
			end: J1{pc: 0xff},
		},
		{
			ins: []Instruction{Call(0xff)},
			end: J1{pc: 0xff, rstack: [0x20]uint16{0x00, 0x02}, rsp: 1},
		},
		{
			ins: []Instruction{Lit(0xff)},
			end: J1{pc: 1, st0: 0xff, dsp: 1},
		},
		{
			ins: []Instruction{Lit(0xff), Lit(0xfe)},
			end: J1{pc: 2, st0: 0xfe, dstack: [0x20]uint16{0x00, 0x00, 0xff}, dsp: 2},
		},
		{ // dup
			ins: []Instruction{Lit(0xff), ALU{Opcode: opT, TtoN: true, Ddir: 1}},
			end: J1{pc: 2, st0: 0xff, dstack: [0x20]uint16{0x00, 0x00, 0xff}, dsp: 2},
		},
		{ // over
			ins: []Instruction{Lit(0xaa), Lit(0xbb), ALU{Opcode: opN, TtoN: true, Ddir: 1}},
			end: J1{pc: 3, st0: 0xaa, dstack: [0x20]uint16{0x00, 0x00, 0xaa, 0xbb}, dsp: 3},
		},
		{ // invert
			ins: []Instruction{Lit(0x00ff), ALU{Opcode: opNotT}},
			end: J1{pc: 2, st0: 0xff00, dsp: 1},
		},
		{ // +
			ins: []Instruction{Lit(1), Lit(2), ALU{Opcode: opTplusN, Ddir: -1}},
			end: J1{pc: 3, st0: 3, dsp: 1, dstack: [0x20]uint16{0, 0, 1}},
		},
		{ // swap
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opN, TtoN: true}},
			end: J1{pc: 3, st0: 2, dsp: 2, dstack: [0x20]uint16{0, 0, 3}},
		},
		{ // nip
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opT, Ddir: -1}},
			end: J1{pc: 3, st0: 3, dsp: 1, dstack: [0x20]uint16{0, 0, 2}},
		},
		{ // drop
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opN, Ddir: -1}},
			end: J1{pc: 3, st0: 2, dsp: 1, dstack: [0x20]uint16{0, 0, 2}},
		},
		{ // ;
			ins: []Instruction{Call(10), Call(20), ALU{Opcode: opT, RtoPC: true, Rdir: -1}},
			end: J1{pc: 11, rsp: 1, rstack: [0x20]uint16{0, 2, 22}},
		},
		{ // >r
			ins: []Instruction{Lit(10), ALU{Opcode: opN, TtoR: true, Ddir: -1, Rdir: 1}},
			end: J1{pc: 2, rsp: 1, rstack: [0x20]uint16{0, 10}},
		},
		{ // r>
			ins: []Instruction{Lit(10), Call(20), ALU{Opcode: opR, TtoN: true, TtoR: true, Ddir: 1, Rdir: -1}},
			end: J1{pc: 21, st0: 4, dsp: 2, dstack: [0x20]uint16{0, 0, 10}, rsp: 0, rstack: [0x20]uint16{10, 4}},
		},
		{ // r@
			ins: []Instruction{Lit(10), ALU{Opcode: opR, TtoN: true, TtoR: true, Ddir: 1}},
			end: J1{pc: 2, dsp: 2, dstack: [0x20]uint16{0, 0, 10}, rstack: [0x20]uint16{10}},
		},
		{ // @
			ins: []Instruction{ALU{Opcode: opAtT}},
			end: J1{pc: 1},
		},
		{ // !
			ins: []Instruction{Lit(1), Lit(0), ALU{Opcode: opN, NtoAtT: true, Ddir: -1}},
			end: J1{pc: 3, st0: 1, dsp: 1, dstack: [0x20]uint16{0, 0, 1}, memory: [0x4000]uint16{1}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			state := New()
			for _, ins := range tc.ins {
				state.eval(ins)
			}
			cmp(t, *state, tc.end)
		})
	}
}

func TestNextST0(t *testing.T) {
	testCases := []struct {
		ins   ALU
		st0   uint16
		state J1
	}{
		{ins: ALU{Opcode: opT}, st0: 0xff, state: J1{st0: 0xff}},
		{ins: ALU{Opcode: opN}, st0: 0xbb, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opTplusN}, st0: 0x01ba, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opTandN}, st0: 0xbb, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opTorN}, st0: 0xff, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opTxorN}, st0: 0x44, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opNotT}, st0: 0xff55, state: J1{st0: 0xaa}},
		{ins: ALU{Opcode: opNeqT}, st0: 0x00, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opNeqT}, st0: 0xffff, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xff}, dsp: 2}},
		{ins: ALU{Opcode: opNleT}, st0: 0xffff, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opNleT}, st0: 0x00, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xff}, dsp: 2}},
		{ins: ALU{Opcode: opNrshiftT}, st0: 0x3f, state: J1{st0: 0x02, dstack: [0x20]uint16{0, 0xaa, 0xff}, dsp: 2}},
		{ins: ALU{Opcode: opTminus1}, st0: 0x54, state: J1{st0: 0x55}},
		{ins: ALU{Opcode: opR}, st0: 0x5, state: J1{rstack: [0x20]uint16{0, 0x05}, rsp: 1}},
		{ins: ALU{Opcode: opAtT}, st0: 0x5, state: J1{st0: 0x02, memory: [0x4000]uint16{0, 5, 10}}},
		{ins: ALU{Opcode: opNlshiftT}, st0: 0x3fc, state: J1{st0: 0x02, dstack: [0x20]uint16{0, 0xaa, 0xff}, dsp: 2}},
		{ins: ALU{Opcode: opDepth}, st0: 0x305, state: J1{rsp: 3, dsp: 5}},
		{ins: ALU{Opcode: opNuleT}, st0: 0xffff, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opNuleT}, st0: 0x00, state: J1{st0: 0xff, dstack: [0x20]uint16{0, 0xaa, 0xff}, dsp: 2}},
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
	j1 := New()
	if err := j1.LoadBytes(data); err != nil {
		t.Fatal(err)
	}
	expect := [0x4000]uint16{0x0201, 0x0804}
	if j1.memory != expect {
		t.Errorf("got %v, want %v", j1.memory[:2], expect)
	}
}

func TestRest(t *testing.T) {
	j1 := &J1{pc: 100, dsp: 2, rsp: 3, st0: 5}
	j1.Reset()
	if j1.pc != 0 || j1.dsp != 0 || j1.rsp != 0 || j1.st0 != 0 {
		t.Errorf("got %v", j1)
	}
}
