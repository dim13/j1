package main

type J1 struct {
	dsp    uint16       // 5 bit Data stack pointer
	st0    uint16       // 5 bit Return stack pointer
	pc     uint16       // 13 bit
	rsp    uint16       // 5 bit
	dstack [0x20]uint16 // Data stack
	rstack [0x20]uint16 // Return stack
	memory [0x8000]uint16
}

func (vm *J1) Eval() {
	insn := vm.memory[vm.pc]
	immediate := insn & 0x7fff // 0,insn[14:0]

	var st0sel uint16

	switch insn & 0x6000 { // insn[14:13]
	case 0x0000: // ubranch
		st0sel = 0
	case 0x4000: // call
		st0sel = 0
	case 0x2000: // 0branch
		st0sel = 1
	case 0x6000: // ALU
		st0sel = (insn >> 8) & 0x0f
	}

	st0 := vm.st0
	st1 := vm.dstack[vm.dsp&0x001f]
	rst0 := vm.rstack[vm.rsp&0x001f]

	//is_alu := insn&0xe000 == 0x6000
	is_lit := insn&0x8000 != 0

	var _st0 uint16
	if is_lit {
		_st0 = immediate
	} else {
		switch st0sel {
		case 0x00: // T
			_st0 = st0
		case 0x01: // N
			_st0 = st1
		case 0x02: // T+N
			_st0 = st0 + st1
		case 0x03: // T&N
			_st0 = st0 & st1
		case 0x04: // T|N
			_st0 = st0 | st1
		case 0x05: // T^N
			_st0 = st0 ^ st1
		case 0x06: // ~T
			_st0 = ^st0
		case 0x07: // N==T
			if st1 == st0 {
				_st0 = 1
			} else {
				_st0 = 0
			}
		case 0x08: // N<T
			if int16(st1) < int16(st0) {
				_st0 = 1
			} else {
				_st0 = 0
			}
		case 0x09: // N>>T
			_st0 = st1 >> (st0 & 0xf)
		case 0x0a: // T-1
			_st0 = st0 - 1
		case 0x0b: // R (rT)
			_st0 = rst0
		case 0x0c: // [T] TODO
			if st0&0xc000 != 0 {
				// io_din
			} else {
				// ramrd
			}
		case 0x0d: // N<<T
			_st0 = st1 << (st0 & 0xf)
		case 0x0e: // depth (dsp)
			_st0 = (vm.rsp << 8) | vm.dsp
		case 0x0f: // Nu<T
			if st1 < st0 {
				_st0 = 1
			} else {
				_st0 = 0
			}
		}
	}

	vm.st0 = _st0
}
