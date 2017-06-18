package j1

import "testing"

func TestBaseWords(t *testing.T) {
	for word, alu := range BaseWords {
		t.Run(word, func(t *testing.T) {
			buf := make([]uint16, len(alu))
			for i, ins := range alu {
				buf[i] = ins.Compile()
			}
			t.Logf("%4.0X", buf)
		})
	}
}
