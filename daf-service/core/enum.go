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
	AC level = iota + 1
	TC
	PC
	CC
	SC
	MC
)

var CLINIC = [5]level{AC, TC, PC, CC, SC}

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
