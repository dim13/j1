package j1

import (
	"fmt"
	"testing"
)

func TestEval(t *testing.T) {
	testCases := []struct {
		begin, end J1
		ins        Instruction
	}{
		{ins: ALU{}, begin: J1{}, end: J1{pc: 1}},
		{ins: Jump(0xff), begin: J1{}, end: J1{pc: 0xff}},
		{ins: Cond(0xff), begin: J1{st0: 1, dsp: 1}, end: J1{pc: 1}},
		{ins: Cond(0xff), begin: J1{st0: 0, dsp: 1}, end: J1{pc: 0xff}},
		{ins: Call(0xff), begin: J1{}, end: J1{pc: 0xff, rstack: [32]uint16{1}, rsp: 1}},
		{ins: Lit(0xff), begin: J1{}, end: J1{pc: 1, st0: 0xff, dstack: [32]uint16{0xff}, dsp: 1}},
		{ins: Lit(0xfe),
			begin: J1{pc: 1, st0: 0xff, dstack: [32]uint16{0xff}, dsp: 1},
			end:   J1{pc: 2, st0: 0xfe, dstack: [32]uint16{0xff, 0xfe}, dsp: 2}},
		{ins: ALU{Opcode: 0, TtoN: true, Ddir: 1}, // dup
			begin: J1{pc: 1, st0: 0xaa, dstack: [32]uint16{0xbb}, dsp: 1},
			end:   J1{pc: 2, st0: 0xaa, dstack: [32]uint16{0xbb, 0xaa}, dsp: 2}},
		{ins: ALU{Opcode: 1, TtoN: true, Ddir: 1}, // over
			begin: J1{pc: 1, st0: 0xaa, dstack: [32]uint16{0xbb}, dsp: 1},
			end:   J1{pc: 2, st0: 0xbb, dstack: [32]uint16{0xbb, 0xaa}, dsp: 2}},
		// TODO
		// ALU{Opcode: 6} // invert
		// ALU{Opcode: 2, Ddir: -1} // +
		// ALU{Opcode: 1, TtoN: true} // swap
		// ALU{Opcode: 0, Ddir: -1} // nip
		// ALU{Opcode: 1, Ddir: -1} // drop
		// ALU{Opcode: 0, RtoPC: true, Rdir: -1} // ;
		// ALU{Opcode: 1, TtoR: true, Ddir: -1, Rdir: 1} // >r
		// ALU{Opcode: 11, TtoN: true, TtoR: true, Ddir: 1, Rdir: -1} // r>
		// ALU{Opcode: 11, TtoN: true, TtoR: true, Ddir: 1} // r@
		// ALU{Opcode: 12} // @
		// ALU{Opcode: 1, Ddir: -1} // !
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			if _, ok := tc.ins.(ALU); ok {
				t.SkipNow()
			}
			state := &tc.begin
			state.eval(tc.ins)
			if *state != tc.end {
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
	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			if tc.state.dsp > 0 {
				t.SkipNow()
			}
			state := &tc.state
			st0 := state.newST0(tc.ins)
			if st0 != tc.st0 {
				t.Errorf("got %x, want %x", st0, tc.st0)
			}
		})
	}
}
