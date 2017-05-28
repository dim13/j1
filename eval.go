package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type J1 struct {
	dsp    uint16       // 5 bit Data stack pointer
	st0    uint16       // 5 bit Return stack pointer
	pc     uint16       // 13 bit
	rsp    uint16       // 5 bit
	dstack [0x20]uint16 // Data stack
	rstack [0x20]uint16 // Return stack
	memory [0x8000]uint16
}

func (vm *J1) String() string {
	s := fmt.Sprintf("\tPC %0.4X\n", vm.pc)
	s += fmt.Sprintf("\tD %v %0.4X %0.4X\n", vm.dsp, vm.dstack[:vm.dsp], vm.st0)
	s += fmt.Sprintf("\tR %v %0.4X\n", vm.rsp, vm.rstack[:vm.rsp])
	return s
}

func (vm *J1) ReadFile(fname string) error {
	fd, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer fd.Close()
	stat, err := fd.Stat()
	if err != nil {
		return err
	}
	size := stat.Size() >> 1
	err = binary.Read(fd, binary.BigEndian, vm.memory[:size])
	if err != io.ErrUnexpectedEOF {
		return err
	}
	return nil
}

func (vm *J1) Eval() {
	ins := Decode(vm.memory[vm.pc])
	switch v := ins.(type) {
	case Lit:
		vm.dstack[vm.dsp] = vm.st0
		vm.dsp += 1
		vm.st0 = uint16(v)
		vm.pc += 1
		fmt.Println(v)
	case Jump:
		vm.pc = uint16(v)
		fmt.Println(v)
	case Cond:
		vm.pc = uint16(v)
		fmt.Println(v)
	case Call:
		vm.rstack[vm.rsp] = vm.pc + 1
		vm.rsp += 1
		vm.pc = uint16(v)
		fmt.Println(v)
	case ALU:
		vm.st0 = vm.ST0(v)
		vm.dsp = uint16(int8(vm.dsp) + v.Ddir)
		vm.rsp = uint16(int8(vm.rsp) + v.Rdir)
		if v.RtoPC {
			vm.pc = vm.R()
		} else {
			vm.pc += 1
		}
		fmt.Println(v)
	}
	fmt.Println(vm)
}

func (vm *J1) T() uint16 { return vm.st0 }
func (vm *J1) N() uint16 { return vm.dstack[vm.dsp] }
func (vm *J1) R() uint16 { return vm.rstack[vm.rsp] }

func (vm *J1) ST0(v ALU) uint16 {
	T := vm.T()
	N := vm.N()
	R := vm.R()
	switch v.Opcode {
	case 0: // T
		return T
	case 1: // N
		return N
	case 2: // T+N
		return T + N
	case 3: // T&N
		return T & N
	case 4: // T|N
		return T | N
	case 5: // T^N
		return T ^ N
	case 6: // ~T
		return ^T
	case 7: // N==T
		if N == T {
			return 1
		}
		return 0
	case 8: // N<T
		if int16(N) < int16(T) {
			return 1
		}
		return 0
	case 9: // N>>T
		return N >> (T & 0xf)
	case 10: // T-1
		return T - 1
	case 11: // R (rT)
		return R
	case 12: // [T]
		return vm.memory[T]
	case 13: // N<<T
		return N << (T & 0xf)
	case 14: // depth (dsp)
		return (vm.rsp << 8) | vm.dsp
	case 15: // Nu<T
		if N < T {
			return 1
		}
		return 0
	}
	return 0
}
