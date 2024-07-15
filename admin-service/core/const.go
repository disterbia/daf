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
	AC level = iota + 1
	TC
	PC
	CC
	SC
	HC
	FC
)

var CLINIC = [7]level{AC, TC, PC, CC, SC, HC, FC}
