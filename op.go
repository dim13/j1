package j1

type Op uint8

const (
	opT Op = iota
	opN
	opTplusN
	opTandN
	opTorN
	opTxorN
	opNotT
	opNeqT
	opNleT
	opNrshiftT
	opTminus1
	opR
	opAtT
	opNlshiftT
	opDepth
	opNuleT
	nOps
)

var opcodeNames = [nOps]string{
	opT:        "T",
	opN:        "N",
	opTplusN:   "T+N",
	opTandN:    "T∧N",
	opTorN:     "T∨N",
	opTxorN:    "T⊻N",
	opNotT:     "¬T",
	opNeqT:     "N=T",
	opNleT:     "N<T",
	opNrshiftT: "N≫T",
	opTminus1:  "T-1",
	opR:        "R",
	opAtT:      "[T]",
	opNlshiftT: "N≪T",
	opDepth:    "D",
	opNuleT:    "Nu<T",
}

func (op Op) String() string {
	return opcodeNames[op]
}
