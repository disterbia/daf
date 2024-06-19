package core

type level int

const (
	ETC level = iota + 1
)

const (
	TR level = iota + 1
	UL
	UR
	LL
	LR
	LOCOMOTION
)
