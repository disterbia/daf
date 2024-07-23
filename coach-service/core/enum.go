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
	WRIST
	FINGER
	HIP
	KNEE
	SUBHIP
	SUBKNEE
	ANKLE
)

const (
	AC level = iota + 1
	TC
	PC
	CC
	SC
)

var CLINIC = [5]level{AC, TC, PC, CC, SC}

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
