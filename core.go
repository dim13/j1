package j1

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

const memSize = 0x4000

type Console interface {
	Read() uint16
	Write(uint16)
	Len() uint16
}

// Core of J1 Forth CPU
type Core struct {
	memory [memSize]uint16 // 0..0x3fff main memory, 0x4000 .. 0x7fff mem-mapped i/o
	pc     uint16          // 13 bit
	st0    uint16          // top of data stack
	d, r   stack           // data and return stacks
	tty    Console         // console i/o
	stop   context.CancelFunc
}

func New() *Core {
	return new(Core)
}

// Reset VM
func (c *Core) Reset() {
	c.pc, c.st0, c.d.sp, c.r.sp = 0, 0, 0, 0
}

// LoadBytes into memory
func (c *Core) LoadBytes(data []byte) error {
	size := len(data) >> 1
	if size >= memSize {
		return fmt.Errorf("too big")
	}
	return binary.Read(bytes.NewReader(data), binary.LittleEndian, c.memory[:size])
}

// LoadFile into memory
func (c *Core) LoadFile(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	return c.LoadBytes(data)
}

// Run evaluates content of memory
func (c *Core) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	c.tty = NewConsole(ctx)
	c.stop = cancel
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.Eval(Decode(c.memory[c.pc]))
		}
	}
}

func (c *Core) String() string {
	s := fmt.Sprintf("\tPC=%0.4X ST=%0.4X\n", c.pc, c.st0)
	s += fmt.Sprintf("\tD=%0.4X\n", c.d.dump())
	s += fmt.Sprintf("\tR=%0.4X\n", c.r.dump())
	return s
}

func (c *Core) writeAt(addr, value uint16) {
	if off := int(addr >> 1); off < memSize {
		c.memory[addr>>1] = value
	}
	switch addr {
	case 0xf000: // key
		c.tty.Write(value)
	case 0xf002: // bye
		c.stop()
	}
}

func (c *Core) readAt(addr uint16) uint16 {
	if off := int(addr >> 1); off < memSize {
		return c.memory[off]
	}
	switch addr {
	case 0xf000: // tx!
		return c.tty.Read()
	case 0xf001: // ?rx
		return c.tty.Len()
	}
	return 0
}

func (c *Core) Eval(ins Instruction) {
	c.pc++
	switch v := ins.(type) {
	case Lit:
		c.d.push(c.st0)
		c.st0 = v.Value()
	case Jump:
		c.pc = v.Value()
	case Call:
		c.r.push(c.pc << 1)
		c.pc = v.Value()
	case Cond:
		if c.st0 == 0 {
			c.pc = v.Value()
		}
		c.st0 = c.d.pop()
	case ALU:
		if v.RtoPC {
			c.pc = c.r.get() >> 1
		}
		if v.NtoAtT {
			c.writeAt(c.st0, c.d.get())
		}
		st0 := c.newST0(v.Opcode)
		c.d.move(v.Ddir)
		c.r.move(v.Rdir)
		if v.TtoN {
			c.d.set(c.st0)
		}
		if v.TtoR {
			c.r.set(c.st0)
		}
		c.st0 = st0
	}
}

func (c *Core) newST0(opcode uint16) uint16 {
	T, N, R := c.st0, c.d.get(), c.r.get()
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
		return c.readAt(T)
	case opNlshiftT: // N<<T
		return N << (T & 0xf)
	case opDepth: // depth (dsp)
		return (c.r.depth() << 8) | c.d.depth()
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
