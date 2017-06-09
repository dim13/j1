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
			ins: []Instruction{Lit(0xff), ALU{Opcode: 0, TtoN: true, Ddir: 1}},
			end: J1{pc: 2, st0: 0xff, dstack: [32]uint16{0x00, 0x00, 0xff}, dsp: 2},
		},
		{ // over
			ins: []Instruction{Lit(0xaa), Lit(0xbb), ALU{Opcode: 1, TtoN: true, Ddir: 1}},
			end: J1{pc: 3, st0: 0xaa, dstack: [32]uint16{0x00, 0x00, 0xaa, 0xbb}, dsp: 3},
		},
		{ // invert
			ins: []Instruction{Lit(0x00ff), ALU{Opcode: 6}},
			end: J1{pc: 2, st0: 0xff00, dsp: 1},
		},
		{ // +
			ins: []Instruction{Lit(1), Lit(2), ALU{Opcode: 2, Ddir: -1}},
			end: J1{pc: 3, st0: 3, dsp: 1, dstack: [32]uint16{0, 0, 1}},
		},
		{ // swap
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: 1, TtoN: true}},
			end: J1{pc: 3, st0: 2, dsp: 2, dstack: [32]uint16{0, 0, 3}},
		},
		{ // nip
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: 0, Ddir: -1}},
			end: J1{pc: 3, st0: 3, dsp: 1, dstack: [32]uint16{0, 0, 2}},
		},
		{ // drop
			ins: []Instruction{Lit(2), Lit(3), ALU{Opcode: 1, Ddir: -1}},
			end: J1{pc: 3, st0: 2, dsp: 1, dstack: [32]uint16{0, 0, 2}},
		},
		{ // ;
			ins: []Instruction{Call(10), Call(20), ALU{Opcode: 0, RtoPC: true, Rdir: -1}},
			end: J1{pc: 11, rsp: 1, rstack: [32]uint16{0, 1, 11}},
		},
		{ // >r
			ins: []Instruction{Lit(10), ALU{Opcode: 1, TtoR: true, Ddir: -1, Rdir: 1}},
			end: J1{pc: 2, rsp: 1, rstack: [32]uint16{0, 10}},
		},
		{ // r>
		//	ins: []Instruction{Lit(10), Call(20), ALU{Opcode: 11, TtoN: true, TtoR: true, Ddir: 1, Rdir: -1}},
		//	end: J1{pc: 21, st0: 2, rsp: 0, rstack: [32]uint16{10, 2}, dsp: 2, dstack: [32]uint16{0, 0, 10}},
		},
		{ // r@
		// ALU{Opcode: 11, TtoN: true, TtoR: true, Ddir: 1}
		},
		{ // @
		// ALU{Opcode: 12}
		},
		{ // !
		// ALU{Opcode: 1, NtoAtT: true, Ddir: -1}
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
		state J1
		st0   uint16
	}{
		{ins: ALU{Opcode: 0}, state: J1{st0: 0xff}, st0: 0xff},
		{ins: ALU{Opcode: 1}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xbb}, dsp: 2}, st0: 0xbb},
		{ins: ALU{Opcode: 2}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xbb}, dsp: 2}, st0: 0x1ba},
		{ins: ALU{Opcode: 3}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xbb}, dsp: 2}, st0: 0xbb},
		{ins: ALU{Opcode: 4}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xbb}, dsp: 2}, st0: 0xff},
		{ins: ALU{Opcode: 5}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xbb}, dsp: 2}, st0: 0x44},
		{ins: ALU{Opcode: 6}, state: J1{st0: 0xaa}, st0: 0xff55},
		{ins: ALU{Opcode: 7}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xbb}, dsp: 2}, st0: 0},
		{ins: ALU{Opcode: 7}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xff}, dsp: 2}, st0: 1},
		{ins: ALU{Opcode: 8}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xbb}, dsp: 2}, st0: 1},
		{ins: ALU{Opcode: 8}, state: J1{st0: 0xff, dstack: [32]uint16{0xaa, 0xff}, dsp: 2}, st0: 0},
		{ins: ALU{Opcode: 9}, state: J1{st0: 0x02, dstack: [32]uint16{0xaa, 0xff}, dsp: 2}, st0: 0x3f},
	}
	t.SkipNow()
	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			state := &tc.state
			st0 := state.newST0(tc.ins)
			if st0 != tc.st0 {
				t.Errorf("got %x, want %x", st0, tc.st0)
			}
		})
	}
}
