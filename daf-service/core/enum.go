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
	NC
	MC
)

const (
	TRUNK level = iota + 1
	SHOULDER
	ELBOW
	WRIST
	FINGER
	HIP
	KNEE
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
