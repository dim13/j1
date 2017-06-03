package j1

import (
	"fmt"
	"testing"
)

func TestEval(t *testing.T) {
	testCases := []struct {
		before J1
		after  J1
		ins    Instruction
	}{
		{ins: ALU{}, before: J1{}, after: J1{pc: 1}},
		{ins: Jump(0xff), before: J1{}, after: J1{pc: 0xff}},
		{ins: Cond(0xff), before: J1{st0: 1, dsp: 1}, after: J1{pc: 1}},
		{ins: Cond(0xff), before: J1{st0: 0, dsp: 1}, after: J1{pc: 0xff}},
		{ins: Call(0xff), before: J1{}, after: J1{pc: 0xff, rstack: [32]uint16{1}, rsp: 1}},
		{ins: Lit(0xff), before: J1{}, after: J1{pc: 1, st0: 0xff, dstack: [32]uint16{0xff}, dsp: 1}},
		{ins: Lit(0xfe),
			before: J1{pc: 1, st0: 0xff, dstack: [32]uint16{0xff}, dsp: 1},
			after:  J1{pc: 2, st0: 0xfe, dstack: [32]uint16{0xff, 0xfe}, dsp: 2}},
		// to be continued
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.ins), func(t *testing.T) {
			state := &tc.before
			state.eval(tc.ins)
			if *state != tc.after {
				t.Errorf("got %v, want %v", state, &tc.after)
			}
		})
	}
}
