package j1

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

const (
	memSize   = 0x4000
	stackSize = 0x20
)

// J1 Forth processor VM
type J1 struct {
	memory  [memSize]uint16 // 0..0x3fff main memory, 0x4000 .. 0x7fff mem-mapped i/o
	pc      uint16          // 13 bit
	st0     uint16          // top of data stack
	d       stack
	r       stack
	console io.ReadWriter
}

func New() *J1 {
	return &J1{console: NewConsole()}
}

// Reset VM
func (j1 *J1) Reset() {
	j1.pc, j1.st0, j1.d.sp, j1.r.sp = 0, 0, 0, 0
}

// LoadBytes into memory
func (j1 *J1) LoadBytes(data []byte) error {
	size := len(data) >> 1
	if size >= memSize {
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
		//fmt.Printf("%4d %v\n%v", n, ins, j1)
	}
}

func (j1 *J1) String() string {
	s := fmt.Sprintf("\tPC=%0.4X ST=%0.4X\n", j1.pc, j1.st0)
	s += fmt.Sprintf("\tD=%0.4X\n", j1.d.dump())
	s += fmt.Sprintf("\tR=%0.4X\n", j1.r.dump())
	return s
}

func (j1 *J1) writeAt(addr, value uint16) {
	if off := int(addr >> 1); off < memSize {
		j1.memory[addr>>1] = value
	}
	switch addr {
	case 0xf000: // key
		fmt.Fprintf(j1.console, "%c", value)
	case 0xf002: // bye
		j1.Reset()
	}
}

func (j1 *J1) readAt(addr uint16) uint16 {
	if off := int(addr >> 1); off < memSize {
		return j1.memory[off]
	}
	switch addr {
	case 0xf000: // tx!
		var b uint16
		fmt.Fscanf(j1.console, "%c", &b)
		return b
	case 0xf001: // ?rx
		return 1
	}
	return 0
}

func (j1 *J1) eval(ins Instruction) {
	j1.pc++
	switch v := ins.(type) {
	case Lit:
		j1.d.push(j1.st0)
		j1.st0 = v.Value()
	case Jump:
		j1.pc = v.Value()
	case Call:
		j1.r.push(j1.pc << 1)
		j1.pc = v.Value()
	case Cond:
		if j1.st0 == 0 {
			j1.pc = v.Value()
		}
		j1.st0 = j1.d.pop()
	case ALU:
		if v.RtoPC {
			j1.pc = j1.r.get() >> 1
		}
		if v.NtoAtT {
			j1.writeAt(j1.st0, j1.d.get())
		}
		st0 := j1.newST0(v.Opcode)
		j1.d.move(v.Ddir)
		j1.r.move(v.Rdir)
		if v.TtoN {
			j1.d.set(j1.st0)
		}
		if v.TtoR {
			j1.r.set(j1.st0)
		}
		j1.st0 = st0
	}
}

func (j1 *J1) newST0(opcode uint16) uint16 {
	T, N, R := j1.st0, j1.d.get(), j1.r.get()
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
		return j1.readAt(T)
	case opNlshiftT: // N<<T
		return N << (T & 0xf)
	case opDepth: // depth (dsp)
		return (j1.r.depth() << 8) | j1.d.depth()
	case opNuleT: // Nu<T
		return bool2int(N < T)
	default:
		panic("invalid instruction")
	}
}

func bool2int(b bool) uint16 {
	if b {
		return ^uint16(0)
	}
	return 0
}
