package j1

import (
	"fmt"
	"testing"
)

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
			end: J1{pc: 0xff, rstack: [32]uint16{0x00, 0x01}, rsp: 1},
		},
		{
			ins: []Instruction{Lit(0xff)},
			end: J1{pc: 1, st0: 0xff, dsp: 1},
		},
		{
			ins: []Instruction{Lit(0xff), Lit(0xfe)},
			end: J1{pc: 2, st0: 0xfe, dstack: [32]uint16{0x00, 0x00, 0xff}, dsp: 2},
		},
		{ // dup
			ins: []Instruction{Lit(0xff), ALU{Opcode: opT, TtoN: true, Ddir: 1}},
			end: J1{pc: 2, st0: 0xff, dstack: [32]uint16{0x00, 0x00, 0xff}, dsp: 2},
		},
		{ // over
			ins: []Instruction{Lit(0xaa), Lit(0xbb), ALU{Opcode: opN, TtoN: true, Ddir: 1}},
			end: J1{pc: 3, st0: 0xaa, dstack: [32]uint16{0x00, 0x00, 0xaa, 0xbb}, dsp: 3},
		},
		{ // invert
			ins: []Instruction{Lit(0x00ff), ALU{Opcode: opNotT}},
			end: J1{pc: 2, st0: 0xff00, dsp: 1},
		},
		{ // +
			ins: []Instruction{Lit(1), Lit(2), ALU{Opcode: opTplusN, Ddir: -1}},
			end: J1{pc: 3, st0: 3, dsp: 1, dstack: [32]uint16{0, 0, 1}},
		},
		{ // swap
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opN, TtoN: true}},
			end: J1{pc: 3, st0: 2, dsp: 2, dstack: [32]uint16{0, 0, 3}},
		},
		{ // nip
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opT, Ddir: -1}},
			end: J1{pc: 3, st0: 3, dsp: 1, dstack: [32]uint16{0, 0, 2}},
		},
		{ // drop
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: opN, Ddir: -1}},
			end: J1{pc: 3, st0: 2, dsp: 1, dstack: [32]uint16{0, 0, 2}},
		},
		{ // ;
			ins: []Instruction{Call(10), Call(20), ALU{Opcode: opT, RtoPC: true, Rdir: -1}},
			end: J1{pc: 11, rsp: 1, rstack: [32]uint16{0, 1, 11}},
		},
		{ // >r
			ins: []Instruction{Lit(10), ALU{Opcode: opN, TtoR: true, Ddir: -1, Rdir: 1}},
			end: J1{pc: 2, rsp: 1, rstack: [32]uint16{0, 10}},
		},
		{ // r>
			ins: []Instruction{Lit(10), Call(20), ALU{Opcode: opR, TtoN: true, TtoR: true, Ddir: 1, Rdir: -1}},
			end: J1{pc: 21, st0: 2, dsp: 2, dstack: [32]uint16{0, 0, 10}, rsp: 0, rstack: [32]uint16{10, 2}},
		},
		{ // r@
			ins: []Instruction{Lit(10), ALU{Opcode: opR, TtoN: true, TtoR: true, Ddir: 1}},
			end: J1{pc: 2, dsp: 2, dstack: [32]uint16{0, 0, 10}, rstack: [32]uint16{10}},
		},
		{ // @
			ins: []Instruction{ALU{Opcode: opAtT}},
			end: J1{pc: 1},
		},
		{ // !
			ins: []Instruction{Lit(1), Lit(0), ALU{Opcode: opN, NtoAtT: true, Ddir: -1}},
			end: J1{pc: 3, st0: 1, dsp: 1, dstack: [32]uint16{0, 0, 1}, memory: [0x8000]uint16{1}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			state := new(J1)
			for _, ins := range tc.ins {
				state.eval(ins)
			}
			if *state != tc.end {
				t.Logf("D=%v", state.dstack)
				t.Logf("R=%v", state.rstack)
				t.Errorf("got %v, want %v", state, &tc.end)
			}
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
		{ins: ALU{Opcode: opN}, st0: 0xbb, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opTplusN}, st0: 0x01ba, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opTandN}, st0: 0xbb, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opTorN}, st0: 0xff, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opTxorN}, st0: 0x44, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opNotT}, st0: 0xff55, state: J1{st0: 0xaa}},
		{ins: ALU{Opcode: opNeqT}, st0: 0x00, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opNeqT}, st0: 0x01, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xff}, dsp: 2}},
		{ins: ALU{Opcode: opNleT}, st0: 0x01, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opNleT}, st0: 0x00, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xff}, dsp: 2}},
		{ins: ALU{Opcode: opNrshiftT}, st0: 0x3f, state: J1{st0: 0x02, dstack: [32]uint16{0, 0xaa, 0xff}, dsp: 2}},
		{ins: ALU{Opcode: opTminus1}, st0: 0x54, state: J1{st0: 0x55}},
		{ins: ALU{Opcode: opR}, st0: 0x5, state: J1{rstack: [32]uint16{0, 0x05}, rsp: 1}},
		{ins: ALU{Opcode: opAtT}, st0: 0x5, state: J1{st0: 0x01, memory: [0x8000]uint16{0, 5, 10}}},
		{ins: ALU{Opcode: opNlshiftT}, st0: 0x3fc, state: J1{st0: 0x02, dstack: [32]uint16{0, 0xaa, 0xff}, dsp: 2}},
		{ins: ALU{Opcode: opDepth}, st0: 0x305, state: J1{rsp: 3, dsp: 5}},
		{ins: ALU{Opcode: opNuleT}, st0: 0x01, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xbb}, dsp: 2}},
		{ins: ALU{Opcode: opNuleT}, st0: 0x00, state: J1{st0: 0xff, dstack: [32]uint16{0, 0xaa, 0xff}, dsp: 2}},
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
	j1 := new(J1)
	if err := j1.LoadBytes(data); err != nil {
		t.Fatal(err)
	}
	expect := [0x8000]uint16{0x0102, 0x0408}
	if j1.memory != expect {
		t.Errorf("got %v, want %v", j1.memory[:2], expect)
	}
}

func TestRest(t *testing.T) {
	vm := &J1{pc: 100, dsp: 2, rsp: 3, st0: 5}
	vm.Reset()
	if vm.pc != 0 || vm.dsp != 0 || vm.rsp != 0 || vm.st0 != 0 {
		t.Errorf("got %v", vm)
	}
}
