package core

type level int

const (
	ABODY level = iota + 1
	UBODY
	LBODY
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

const (
	AC level = iota + 1
	TC
	PC
	CC
	SC
	NC
)

var CLINIC = [6]level{AC, TC, PC, CC, SC, NC}

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
