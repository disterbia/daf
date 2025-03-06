package core

type snsType uint

// 0은 daf에서 가입
const (
	Password snsType = iota + 1
	Kakao
	Google
	Apple
	Facebook
	Naver
)

var profileImageType = 1

var ADAPFIT = 1
var DAF = 2
