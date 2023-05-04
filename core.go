package j1

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
)

// Console i/o
type Console interface {
	Read() uint16
	Write(uint16)
	Len() uint16
	Stop()
}

// Core of J1 Forth CPU
//
//	33 deep × 16 bit data stack
//	32 deep × 16 bit return stack
//	13 bit program counter
//	memory is 16 bit wide and addressed by bytes
//	0..0x3fff RAM, 0x4000..0x7fff mem-mapped I/O
type Core struct {
	memory  [8192]uint16 // 0..0x3fff RAM, 0x4000..0x7fff mem-mapped I/O
	pc      uint16       // 13 bit
	st0     uint16       // top of data stack
	d, r    stack        // data and return stacks
	console Console      // console i/o
}

// New core with console i/o
func New(con Console) *Core {
	return &Core{console: con}
}

// Reset VM
func (c *Core) Reset() {
	c.pc, c.st0, c.d.sp, c.r.sp = 0, 0, 0, 0
}

// Write memory
func (c *Core) Write(data []byte) (int, error) {
	size := len(data) >> 1
	if size >= len(c.memory) {
		return 0, fmt.Errorf("data size %v > memory size %v", size, len(c.memory))
	}
	return len(data), binary.Read(bytes.NewReader(data), binary.LittleEndian, c.memory[:size])
}

func (c *Core) String() string {
	s := fmt.Sprintf("\tPC=%0.4X ST=%0.4X\n", c.pc, c.st0)
	s += fmt.Sprintf("\tD=%0.4X\n", c.d.dump())
	s += fmt.Sprintf("\tR=%0.4X\n", c.r.dump())
	return s
}

const ioMask = 3 << 14

func (c *Core) writeAt(addr, value uint16) {
	if addr&ioMask == 0 {
		c.memory[addr>>1] = value
	}
	switch addr {
	case 0x7000: // key
		c.console.Write(value)
	case 0x7002: // bye
		c.console.Stop()
	}
}

func (c *Core) readAt(addr uint16) uint16 {
	if addr&ioMask == 0 {
		return c.memory[addr>>1]
	}
	switch addr {
	case 0x7000: // tx!
		return c.console.Read()
	case 0x7001: // ?rx
		return c.console.Len()
	}
	return 0
}

// Run evaluates content of memory
func (c *Core) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.Execute(c.Fetch())
		}
	}
}

// Fetch instruction at current program counter position
func (c *Core) Fetch() Instruction {
	return Decode(c.memory[c.pc])
}

// Execute instruction
func (c *Core) Execute(ins Instruction) {
	c.pc++
	switch v := ins.(type) {
	case Literal:
		c.d.push(c.st0)
		c.st0 = v.value()
	case Jump:
		c.pc = v.value()
	case Call:
		c.r.push(c.pc << 1)
		c.pc = v.value()
	case Conditional:
		if c.st0 == 0 {
			c.pc = v.value()
		}
		c.st0 = c.d.pop()
	case ALU:
		if v.RtoPC {
			c.pc = c.r.peek() >> 1
		}
		if v.NtoAtT {
			c.writeAt(c.st0, c.d.peek())
		}
		st0 := c.newST0(v.Opcode)
		c.d.move(v.Ddir)
		c.r.move(v.Rdir)
		if v.TtoN {
			c.d.replace(c.st0)
		}
		if v.TtoR {
			c.r.replace(c.st0)
		}
		c.st0 = st0
	}
}

var boolValue = map[bool]uint16{
	false: 0,
	true:  ^uint16(0),
}

func (c *Core) newST0(opcode uint16) uint16 {
	T, N, R := c.st0, c.d.peek(), c.r.peek()
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
		return boolValue[N == T]
	case opNleT: // N<T
		return boolValue[int16(N) < int16(T)]
	case opNrshiftT: // N>>T
		return N >> (T & 0xf)
	case opTminus1: // T-1
		return T - 1
	case opR: // R (rT)
		return R
	case opAtT: // [T]
		return c.readAt(T)
	case opNlshiftT: // N<<T
		return N << (T & 0xf)
	case opDepth: // depth (dsp)
		return (c.r.depth() << 8) | c.d.depth()
	case opNuleT: // Nu<T
		return boolValue[N < T]
	default:
		panic("invalid instruction")
	}
}
