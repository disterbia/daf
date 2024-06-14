package core

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	AgencyId    uint   `json:"agency_id"`
	Name        string `json:"name"`
	EnglishName string `json:"english_name"`
	Phone       string `json:"phone" example:"01000000000"`
	Tel         string `json:"tel" example:"0510000000"`
	Fax         string `json:"fax" examlple:"000000000"`
}

type LoginResponse struct {
	Jwt string `json:"jwt,omitempty"`
	Err string `json:"err,omitempty"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code" example:"인증번호 6자리"`
}

type SaveUserRequest struct {
	ID               uint   `json:"id"`
	Uid              uint   `json:"-"`
	Name             string `json:"name"`
	Gender           bool   `json:"gender"`
	Birthday         string `json:"birthday" example:"yyyy-mm-dd"`
	Phone            string `json:"phone"`
	Addr             string `json:"addr"`
	AddrDetail       string `json:"addr_detail"`
	Memo             string `json:"memo"`
	AgencyID         uint   `json:"agency_id"`
	AdminID          uint   `json:"admin_id"`
	RegistDay        string `json:"regist_day" example:"yyyy-mm-dd"`
	UseStatusID      uint   `json:"use_status_id"`
	DisableTypeIDs   []uint `json:"disable_type_ids"`
	VisitPurposeIDs  []uint `json:"visit_purpose_ids"`
	DisableDetailIDs []uint `json:"disable_detail_ids"`
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
