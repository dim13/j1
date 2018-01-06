package j1

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Console struct {
	io.Reader
	io.Writer
}

// J1 Forth processor VM
type J1 struct {
	pc      uint16         // 13 bit
	st0     uint16         // top of data stack
	dsp     int8           // 5 bit data stack pointer
	rsp     int8           // 5 bit retrun stack pointer
	dstack  [0x20]uint16   // data stack
	rstack  [0x20]uint16   // return stack
	memory  [0x8000]uint16 // 0..0x3fff main memory, 0x4000 .. 0x7fff mem-mapped i/o
	console io.ReadWriter
}

func New() *J1 {
	return &J1{console: Console{Reader: os.Stdin, Writer: os.Stdout}}
}

// Reset VM
func (j1 *J1) Reset() {
	j1.pc, j1.st0, j1.dsp, j1.rsp = 0, 0, 0, 0
}

// LoadBytes into memory
func (j1 *J1) LoadBytes(data []byte) error {
	size := len(data) >> 1
	if size > len(j1.memory) {
		return fmt.Errorf("too big")
	}
	return binary.Read(bytes.NewReader(data), binary.LittleEndian, j1.memory[:size])
}

// LoadFile into memory
func (j1 *J1) LoadFile(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	return j1.LoadBytes(data)
}

// Eval evaluates content of memory
func (j1 *J1) Eval() {
	for n := 0; ; n++ {
		ins := Decode(j1.memory[j1.pc])
		if ins == Jump(0) {
			return
		}
		j1.eval(ins)
		fmt.Printf("%4d %v\n%v\n", n, ins, j1)
	}
}

func (j1 *J1) String() string {
	var rstack [32]uint16
	for i, v := range j1.rstack {
		rstack[i] = v << 1
	}
	s := fmt.Sprintf("\tPC=%0.4X ST=%0.4X\n", j1.pc<<1, j1.st0)
	s += fmt.Sprintf("\tD=%0.4X\n", j1.dstack[:j1.dsp+1])
	s += fmt.Sprintf("\tR=%0.4X\n", rstack[:j1.rsp+1])
	return s
}

func (j1 *J1) eval(ins Instruction) {
	j1.pc++
	switch v := ins.(type) {
	case Lit:
		j1.dsp++
		j1.dstack[j1.dsp] = j1.st0
		j1.st0 = v.Value()
	case Jump:
		j1.pc = v.Value()
	case Call:
		j1.rsp++
		j1.rstack[j1.rsp] = j1.pc
		j1.pc = v.Value()
	case Cond:
		if j1.st0 == 0 {
			j1.pc = v.Value()
		}
		j1.st0 = j1.dstack[j1.dsp] // N
		j1.dsp--
	case ALU:
		if v.RtoPC {
			j1.pc = j1.rstack[j1.rsp]
		}
		st0 := j1.newST0(v.Opcode)
		j1.dsp += v.Ddir
		j1.rsp += v.Rdir
		if v.TtoN {
			j1.dstack[j1.dsp] = j1.st0
		}
		if v.TtoR {
			j1.rstack[j1.rsp] = j1.st0
		}
		if v.NtoAtT {
			switch j1.st0 {
			case 0xf000: // key
				fmt.Fprintf(j1.console, "%c", j1.dstack[j1.dsp])
			case 0xf002: // bye
				j1.rsp = 0
			default:
				j1.memory[j1.st0] = j1.dstack[j1.dsp]
			}
		}
		j1.st0 = st0
	}
}

func bool2int(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}

func (j1 *J1) newST0(opcode uint16) uint16 {
	T, N, R := j1.st0, j1.dstack[j1.dsp], j1.rstack[j1.rsp]
	switch opcode {
	case opT: // T
		return T
	case opN: // N
		return N
	case opTplusN: // T+N
		return T + N
	case opTandN: // T&N
		return T & N
	case opTorN: // T|N
		return T | N
	case opTxorN: // T^N
		return T ^ N
	case opNotT: // ~T
		return ^T
	case opNeqT: // N==T
		return bool2int(N == T)
	case opNleT: // N<T
		return bool2int(int16(N) < int16(T))
	case opNrshiftT: // N>>T
		return N >> (T & 0xf)
	case opTminus1: // T-1
		return T - 1
	case opR: // R (rT)
		return R
	case opAtT: // [T]
		switch T {
		case 0xf000: // tx!
			var b uint16
			fmt.Fscanf(j1.console, "%c", &b)
			return b
		case 0xf001: // ?rx
			return 1
		default:
			return j1.memory[T]
		}
	case opNlshiftT: // N<<T
		return N << (T & 0xf)
	case opDepth: // depth (dsp)
		return (uint16(j1.rsp) << 8) | uint16(j1.dsp)
	case opNuleT: // Nu<T
		return bool2int(N < T)
	default:
		panic("invalid instruction")
	}
}
