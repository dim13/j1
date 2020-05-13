package j1

import "testing"

func TestStack(t *testing.T) {
	s := new(stack)
	s.push(1)
	s.push(2)
	s.push(3)
	if v := s.depth(); v != 3 {
		t.Errorf("depth: got %v, want 3", v)
	}
	if v := s.peek(); v != 3 {
		t.Errorf("peek: got %v, want 3", v)
	}
	s.replace(4)
	if v := s.pop(); v != 4 {
		t.Errorf("pop: got %v, want 4", v)
	}
	if v := s.dump(); v[0] != 1 || v[1] != 2 {
		t.Errorf("pop: got %v, want [1, 2]", v)
	}
}
