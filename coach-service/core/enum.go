package core

type level int

const (
	TBODY level = iota + 1
	UBODY
	LBODY
	LOCOBODY
)

const (
	SHOULDER level = iota + 1
	ELBOW
	HIP
	KNEE
	HANDS
	FOOT
)

const (
	AC level = iota + 1
	TC
	PC
	CC
	SC
	HC
	FC
)

var CLINIC = [7]level{AC, TC, PC, CC, SC, HC, FC}

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
