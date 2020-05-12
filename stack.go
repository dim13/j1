package j1

type stack struct {
	data [32]uint16 // stack
	sp   int8       // 5 bit stack pointer
}

func (s *stack) move(dir int8) {
	s.sp = (s.sp + dir) & 0x1f
}

func (s *stack) push(v uint16) {
	s.sp = (s.sp + 1) & 0x1f
	s.data[s.sp] = v
}

func (s *stack) pop() uint16 {
	sp := s.sp
	s.sp = (s.sp - 1) & 0x1f
	return s.data[sp]
}

func (s *stack) peek() uint16 {
	return s.data[s.sp]
}

func (s *stack) replace(v uint16) {
	s.data[s.sp] = v
}

func (s *stack) depth() uint16 {
	return uint16(s.sp)
}

func (s *stack) dump() []uint16 {
	return s.data[:s.sp+1]
}
