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
	NC level = iota + 1
	TC
	PC
	CC
	SC
	AC
)

var CLINIC = [6]level{NC, TC, PC, CC, SC, AC}
