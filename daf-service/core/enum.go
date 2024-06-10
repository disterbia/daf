package core

type level int

const (
	UL level = iota + 1
	UR
	LL
	LR
	TR
	LOCO
)

const (
	TBODY level = iota + 1
	UBODY
	LBODY
	LOCOBODY
)

const DafCount = 5

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
