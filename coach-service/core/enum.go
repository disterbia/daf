package core

type level int

const (
	TBODY level = iota + 1
	UBODY
	LBODY
)

const (
	NC level = iota + 1
	PC
	CC
	SC
	WC
	AC
	TC
)

var CLINIC = [7]level{NC, PC, CC, SC, WC, AC, TC}

// const (
// 	ULS level = iota + 1
// 	ULE
// 	ULW
// 	ULF

// 	URS
// 	URE
// 	URW
// 	URF

// 	LLH
// 	LLK
// 	LLA
// 	LLT

// 	LRH
// 	LRK
// 	LRA
// 	LRT

// 	TR
// )
