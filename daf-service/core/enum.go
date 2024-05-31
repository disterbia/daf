package core

type level int

const (
	UL level = iota + 1
	UR
	LL
	LR
	TR
)

const (
	TBODY level = iota + 1
	UBODY
	LBODY
)

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
