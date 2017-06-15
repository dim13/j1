package j1

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"time"
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

// Reset VM
func (vm *J1) Reset() {
	vm.dsp = 0
	vm.st0 = 0
	vm.pc = 0
	vm.rsp = 0
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
	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()
	for range ticker.C {
		ins := Decode(vm.memory[vm.pc])
		if ins == Jump(0) {
			return
		}
		vm.eval(ins)
		fmt.Println(ins)
		fmt.Println(vm)
	}
}

func (vm *J1) String() string {
	var rstack [32]uint16
	for i, v := range vm.rstack {
		rstack[i] = v << 1
	}
	s := fmt.Sprintf("\tPC=%0.4X ST=%0.4X\n", vm.pc<<1, vm.st0)
	s += fmt.Sprintf("\tD=%0.4X\n", vm.dstack[:vm.dsp+1])
	s += fmt.Sprintf("\tR=%0.4X\n", rstack[:vm.rsp+1])
	return s
}

func (vm *J1) eval(ins Instruction) {
	switch v := ins.(type) {
	case Lit:
		vm.pc++
		vm.dsp++
		vm.dstack[vm.dsp] = vm.st0
		vm.st0 = uint16(v)
	case Jump:
		vm.pc = uint16(v)
	case Call:
		vm.rsp++
		vm.rstack[vm.rsp] = vm.pc + 1
		vm.pc = uint16(v)
	case Cond:
		vm.pc++
		if vm.st0 == 0 {
			vm.pc = uint16(v)
		}
		vm.st0 = vm.dstack[vm.dsp] // N
		vm.dsp--
	case ALU:
		st0 := vm.newST0(v)
		vm.pc++
		if v.RtoPC {
			vm.pc = vm.rstack[vm.rsp]
		}
		if v.NtoAtT {
			vm.memory[vm.st0] = vm.dstack[vm.dsp]
		}
		vm.dsp = uint16(int8(vm.dsp) + v.Ddir)
		vm.rsp = uint16(int8(vm.rsp) + v.Rdir)
		if v.TtoR {
			vm.rstack[vm.rsp] = vm.st0
		}
		if v.TtoN {
			vm.dstack[vm.dsp] = vm.st0
		}
		vm.st0 = st0
	}
}

func (vm *J1) newST0(v ALU) uint16 {
	T, N, R := vm.st0, vm.dstack[vm.dsp], vm.rstack[vm.rsp]
	switch v.Opcode {
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
		if N == T {
			return 1
		}
		return 0
	case opNleT: // N<T
		if int16(N) < int16(T) {
			return 1
		}
		return 0
	case opNrshiftT: // N>>T
		return N >> (T & 0xf)
	case opTminus1: // T-1
		return T - 1
	case opR: // R (rT)
		return R
	case opAtT: // [T]
		return vm.memory[T]
	case opNlshiftT: // N<<T
		return N << (T & 0xf)
	case opDepth: // depth (dsp)
		return (vm.rsp << 8) | vm.dsp
	case opNuleT: // Nu<T
		if N < T {
			return 1
		}
		return 0
	default:
		panic("invalid instruction")
	}
}
