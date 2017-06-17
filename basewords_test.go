package j1

import "testing"

func TestBaseWords(t *testing.T) {
	for word, alu := range BaseWords {
		t.Run(word, func(t *testing.T) {
			for _, ins := range alu {
				t.Logf("%4.0X", ins.Compile())
			}
		})
	}
}
