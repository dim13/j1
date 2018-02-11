package j1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
)

var ErrStop = errors.New("stop")

// Console i/o
type Console interface {
	Read() uint16
	Write(uint16)
	Len() uint16
}

// Core of J1 Forth CPU
//
// memory: 0x2000 words (16k) addressed by byte
type Core struct {
	memory  [0x2000]uint16 // 0..0x3fff main memory, 0x4000 .. 0xffff mem-mapped i/o
	pc      uint16         // 13 bit
	st0     uint16         // top of data stack
	d, r    stack          // data and return stacks
	console Console        // console i/o
}

// New core with console i/o
func New(con Console) *Core {
	return &Core{console: con}
}

// Reset VM
func (c *Core) Reset() {
	c.pc, c.st0, c.d.sp, c.r.sp = 0, 0, 0, 0
}

// LoadBytes into memory
func (c *Core) LoadBytes(data []byte) error {
	size := len(data) >> 1
	if size >= len(c.memory) {
		return errors.New("too big")
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

func (c *Core) String() string {
	s := fmt.Sprintf("\tPC=%0.4X ST=%0.4X\n", c.pc, c.st0)
	s += fmt.Sprintf("\tD=%0.4X\n", c.d.dump())
	s += fmt.Sprintf("\tR=%0.4X\n", c.r.dump())
	return s
}

const ioMask = 3 << 14

func (c *Core) writeAt(addr, value uint16) error {
	if addr&ioMask == 0 {
		c.memory[addr>>1] = value
	}
	switch addr {
	case 0xf000: // key
		c.console.Write(value)
	case 0xf002: // bye
		return ErrStop
	}
	return nil
}

func (c *Core) readAt(addr uint16) uint16 {
	if addr&ioMask == 0 {
		return c.memory[addr>>1]
	}
	switch addr {
	case 0xf000: // tx!
		return c.console.Read()
	case 0xf001: // ?rx
		return c.console.Len()
	}
	return 0
}

// Run evaluates content of memory
func (c *Core) Run() {
	for {
		ins := c.Decode()
		err := c.Eval(ins)
		if err == ErrStop {
			return
		}
	}
}

// Decode instruction
func (c *Core) Decode() Instruction {
	return Decode(c.memory[c.pc])
}

// Eval instruction
func (c *Core) Eval(ins Instruction) error {
	c.pc++
	switch v := ins.(type) {
	case Lit:
		c.d.push(c.st0)
		c.st0 = v.value()
	case Jump:
		c.pc = v.value()
	case Call:
		c.r.push(c.pc << 1)
		c.pc = v.value()
	case Cond:
		if c.st0 == 0 {
			c.pc = v.value()
		}
		c.st0 = c.d.pop()
	case ALU:
		if v.RtoPC {
			c.pc = c.r.peek() >> 1
		}
		if v.NtoAtT {
			err := c.writeAt(c.st0, c.d.peek())
			if err != nil {
				return err
			}
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
	return nil
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
