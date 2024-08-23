package core

type level int

const (
	NONE level = iota + 1
	ETC
)

const (
	TR level = iota + 1
	UL
	UR
	LL
	LR
	LOCOMOTION
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
	MC
)
