package j1

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

// J1 Forth processor VM
type J1 struct {
	dsp    uint16         // 5 bit data stack pointer
	st0    uint16         // top of data stack
	pc     uint16         // 13 bit
	rsp    uint16         // 5 bit retrun stack pointer
	dstack [0x20]uint16   // data stack
	rstack [0x20]uint16   // deturn stack
	memory [0x8000]uint16 // memory
}

func (vm *J1) String() string {
	return fmt.Sprintf("PC=%0.4X ST=%0.4X D=%0.4X R=%0.4X",
		vm.pc, vm.st0, vm.dstack[:vm.dsp], vm.rstack[:vm.rsp])
}

// LoadBytes into memory
func (vm *J1) LoadBytes(data []byte) error {
	size := len(data) >> 1
	if size > len(vm.memory) {
		return fmt.Errorf("too big")
	}
	return binary.Read(bytes.NewReader(data), binary.BigEndian, vm.memory[:size])
}

// LoadFile into memory
func (vm *J1) LoadFile(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	return vm.LoadBytes(data)
}

// Eval evaluates content of memory
func (vm *J1) Eval() {
	for {
		ins := Decode(vm.memory[vm.pc])
		if ins == Jump(0) {
			break
		}
		vm.eval(ins)
		fmt.Printf("%v\t%v\n\n", ins, vm)
	}
}

var (
	opT = ALU{Opcode: 0}
	opN = ALU{Opcode: 1}
)

func (vm *J1) eval(ins Instruction) {
	dsp := vm.dsp
	pc := vm.pc + 1
	st0 := vm.st0
	rsp := vm.rsp

	switch v := ins.(type) {
	case Lit:
		st0 = uint16(v)
		dsp = vm.dsp + 1
		vm.dstack[vm.dsp] = vm.T()
	case Jump:
		st0 = vm.newST0(opT)
		pc = uint16(v)
	case Call:
		st0 = vm.newST0(opT)
		rsp = vm.rsp + 1
		vm.rstack[vm.rsp] = pc
		pc = uint16(v)
	case Cond:
		st0 = vm.newST0(opN)
		dsp = vm.dsp - 1
		if vm.st0 == 0 {
			pc = uint16(v)
		}
	case ALU:
		st0 = vm.newST0(v)
		if v.RtoPC {
			pc = vm.R()
		}
		if v.NtoAtT {
			vm.memory[vm.st0] = vm.N()
		}
		dsp = uint16(int8(vm.dsp)+v.Ddir) % 32
		rsp = uint16(int8(vm.rsp)+v.Rdir) % 32
		if v.TtoR {
			vm.rstack[(rsp-1)%32] = vm.T()
		}
		if v.TtoN {
			vm.dstack[(dsp-1)%32] = vm.T()
		}
	}

	vm.dsp = dsp
	vm.pc = pc
	vm.st0 = st0
	vm.rsp = rsp
}

// T is top of data stack
func (vm *J1) T() uint16 { return vm.st0 }

// N is second element of data stack
func (vm *J1) N() uint16 { return vm.dstack[(vm.dsp-1)%32] }

// R is top of return stack
func (vm *J1) R() uint16 { return vm.rstack[(vm.rsp-1)%32] }

func (vm *J1) newST0(v ALU) uint16 {
	switch v.Opcode {
	case 0: // T
		return vm.T()
	case 1: // N
		return vm.N()
	case 2: // T+N
		return vm.T() + vm.N()
	case 3: // T&N
		return vm.T() & vm.N()
	case 4: // T|N
		return vm.T() | vm.N()
	case 5: // T^N
		return vm.T() ^ vm.N()
	case 6: // ~T
		return ^vm.T()
	case 7: // N==T
		if vm.N() == vm.T() {
			return 1
		}
		return 0
	case 8: // N<T
		if int16(vm.N()) < int16(vm.T()) {
			return 1
		}
		return 0
	case 9: // N>>T
		return vm.N() >> (vm.T() & 0xf)
	case 10: // T-1
		return vm.T() - 1
	case 11: // R (rT)
		return vm.R()
	case 12: // [T]
		return vm.memory[vm.T()]
	case 13: // N<<T
		return vm.N() << (vm.T() & 0xf)
	case 14: // depth (dsp)
		return (vm.rsp << 8) | vm.dsp
	case 15: // Nu<T
		if vm.N() < vm.T() {
			return 1
		}
		return 0
	default:
		panic("invalid instruction")
	}
}
