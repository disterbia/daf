package core

type UserJointActionRequest struct {
	ID  uint `json:"-"`
	ULS string
	ULE string
	ULW string
	ULF string

	URS string
	URE string
	URW string
	URF string

	LLH string
	LLK string
	LLA string
	LLT string

	LRH string
	LRK string
	LRA string
	LRT string

	TR string
}

type UserJointActionResponse struct {
	ULS  string
	ULE  string
	ULW  string
	ULF  string
	ULAV string

	URS  string
	URE  string
	URW  string
	URF  string
	URAV string

	LLH  string
	LLK  string
	LLA  string
	LLT  string
	LLAV string

	LRH  string
	LRK  string
	LRA  string
	LRT  string
	LRAV string

	TR string
}

type BasicResponse struct {
	Code string `json:"code"`
}

// // for swagger ////
type SuccessResponse struct {
	Jwt string `json:"jwt"`
}
type ErrorResponse struct {
	Err string `json:"err"`
}
