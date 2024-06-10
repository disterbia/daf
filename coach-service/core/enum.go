package core

type level int

const (
	TBODY level = iota + 1
	UBODY
	LBODY
	LOCOBODY
)

const (
	NC level = iota + 1
	PC
	CC
	SC
	AC
	TC
)

var CLINIC = [6]level{NC, PC, CC, SC, AC, TC}

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
