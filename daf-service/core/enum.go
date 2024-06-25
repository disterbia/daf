package core

type level uint

const (
	TR level = iota + 1
	UL
	UR
	LL
	LR
	LOCOMOTION
)

const (
	TBODY level = iota + 1
	UBODY
	LBODY
	LOCOBODY
)

const RECOMMENDCOUNT = 3

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
