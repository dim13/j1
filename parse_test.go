package j1

import (
	"fmt"
	"testing"
)

func TestDecode(t *testing.T) {
	testCases := []struct {
		bin uint16
		ins Instruction
	}{
		{0x0000, Jump(0x0000)},
		{0x1fff, Jump(0x1fff)},
		{0x2000, Cond(0x0000)},
		{0x3fff, Cond(0x1fff)},
		{0x4000, Call(0x0000)},
		{0x5fff, Call(0x1fff)},
		{0x8000, Lit(0x0000)},
		{0xffff, Lit(0x7fff)},
		{0x6000, ALU{Opcode: 0}},
		{0x6100, ALU{Opcode: 1}},
		{0x7000, ALU{Opcode: 0, RtoPC: true}},
		{0x6080, ALU{Opcode: 0, TtoN: true}},
		{0x6040, ALU{Opcode: 0, TtoR: true}},
		{0x6020, ALU{Opcode: 0, NtoAtT: true}},
		{0x600c, ALU{Opcode: 0, Rdir: -1}},
		{0x6004, ALU{Opcode: 0, Rdir: 1}},
		{0x6003, ALU{Opcode: 0, Ddir: -1}},
		{0x6001, ALU{Opcode: 0, Ddir: 1}},
		{0x6f00, ALU{Opcode: 15}},
		{0x70e5, ALU{Opcode: 0, RtoPC: true, TtoN: true, TtoR: true, NtoAtT: true, Rdir: 1, Ddir: 1}},
		{0x7fef, ALU{Opcode: 15, RtoPC: true, TtoN: true, TtoR: true, NtoAtT: true, Rdir: -1, Ddir: -1}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			ins := Decode(tc.bin)
			if ins != tc.ins {
				t.Errorf("got %v, want %v", ins, tc.ins)
			}
			if v := ins.compile(); v != tc.bin {
				t.Errorf("got %v, want %v", v, tc.bin)
			}
		})
	}
}
