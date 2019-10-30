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

type io interface {
	readAt(addr uint16) uint16
	writeAt(addr uint16, value uint16)
}

// Core of J1 Forth CPU
//
//  33 deep × 16 bit data stack
//  32 deep × 16 bit return stack
//  13 bit program counter
//  memory is 16 bit wide and addressed by bytes
//  0..0x3fff RAM, 0x4000..0x7fff mem-mapped I/O
//
type Core struct {
	memory  [0x2000]uint16 // 0..0x3fff RAM, 0x4000..0x7fff mem-mapped I/O
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
	case 0x7000: // key
		c.console.Write(value)
	case 0x7002: // bye
		return ErrStop
	}
	return nil
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
func (c *Core) Run() {
	for {
		ins := c.Fetch()
		err := c.Execute(ins)
		if err == ErrStop {
			return
		}
	}
}

// Fetch instruction at current program counter position
func (c *Core) Fetch() Instruction {
	return Decode(c.memory[c.pc])
}

// Execute instruction
func (c *Core) Execute(ins Instruction) error {
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
		if v.Ret {
			c.pc = c.r.peek() >> 1
		}
		if v.Mod&ModNtoAtT != 0 {
			err := c.writeAt(c.st0, c.d.peek())
			if err != nil {
				return err
			}
		}
		st0 := c.newST0(v.Opcode)
		c.d.move(v.Ddir)
		c.r.move(v.Rdir)
		if v.Mod&ModTtoN != 0 {
			c.d.replace(c.st0)
		}
		if v.Mod&ModTtoR != 0 {
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

func (c *Core) newST0(opcode Opcode) uint16 {
	T, N, R := c.st0, c.d.peek(), c.r.peek()
	switch opcode {
	case OpT: // T
		return T
	case OpN: // N
		return N
	case OpTplusN: // T+N
		return T + N
	case OpTandN: // T&N
		return T & N
	case OpTorN: // T|N
		return T | N
	case OpTxorN: // T^N
		return T ^ N
	case OpNotT: // ~T
		return ^T
	case OpNeqT: // N==T
		return boolValue[N == T]
	case OpNleT: // N<T
		return boolValue[int16(N) < int16(T)]
	case OpNrshiftT: // N>>T
		return N >> (T & 0xf)
	case OpNlshiftT: // N<<T
		return N << (T & 0xf)
	case OpR: // R (rT)
		return R
	case OpAtT: // [T]
		return c.readAt(T)
	case OpIoAtT: // io[T]
		return c.readAt(T)
	case OpStatus: // depth (dsp)
		return (c.r.depth() << 8) | c.d.depth()
	case OpNuLeT: // Nu<T
		return boolValue[N < T]
	default:
		panic("invalid instruction")
	}
}
