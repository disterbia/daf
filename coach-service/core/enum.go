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
	TC
	PC
	CC
	SC
	AC
)

var CLINIC = [6]level{NC, TC, PC, CC, SC, AC}

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
