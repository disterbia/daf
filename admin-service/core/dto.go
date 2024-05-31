package core

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	AgencyId uint   `json:"agency_id"`
}

type LoginResponse struct {
	Jwt string `json:"jwt,omitempty"`
	Err string `json:"err,omitempty"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code" example:"인증번호 6자리"`
}

type ExerciseResponse struct {
	ID   uint
	Name string
}

type CategoryResponse struct {
	ID   uint
	Name string
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
