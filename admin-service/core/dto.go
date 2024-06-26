package core

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	AgencyID    uint   `json:"agency_id"`
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
	Gender           uint   `json:"gender"`
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

type SearchUserRequest struct {
	Page             uint         `json:"page"`
	Name             string       `json:"name"`
	Gender           uint         `json:"gender"`
	AgeCode          uint         `json:"age_code"`
	AgencyID         uint         `json:"ageny_id"`
	AdminID          uint         `json:"admin_id"`
	RegistDay        string       `json:"regist_day" example:"yyyy-mm-dd"`
	UseStatusID      uint         `json:"use_status_id"`
	DisableTypeIDs   []uint       `json:"disable_type_ids"`
	VisitPurposeIDs  []uint       `json:"visit_purpose_ids"`
	DisableDetailIDs []uint       `json:"disable_detail_ids"`
	Afcs             []AfcRequest `json:"afcs"`
}

type SearchUserResponse struct {
	ID             uint             `json:"id"`
	Phone          string           `json:"phone" example:"01000000000"`
	Name           string           `json:"name"`
	Gender         uint             `json:"gender"`
	AgeCode        uint             `json:"age_code"`
	AgencyId       uint             `json:"agency_id"`
	AgencyName     string           `json:"ageny_name"`
	AdminId        uint             `json:"admin_id"`
	AdminName      string           `json:"admin_name"`
	RegistDay      string           `json:"regist_day" example:"yyyy-mm-dd"`
	UseStatusId    uint             `json:"use_status_id"`
	UseStatusName  string           `json:"use_status_name"`
	DisableTypes   []IdNameResponse `json:"disable_types"`
	VisitPurposes  []IdNameResponse `json:"visit_purposes"`
	DisableDetails []IdNameResponse `json:"disable_details"`
	Afc            []AfcResponse    `json:"afc"`
	Addr           string           `json:"addr"`
	Memo           string           `json:"memo"`
	Birthday       string           `json:"birthday" example:"yyyy-mm-dd"`
}

type IdNameResponse struct {
	Id   uint   `json:"id"`
	Name string `json:"string"`
}

type AfcRequest struct {
	BodyCompositionID uint `json:"body_composition_id"`
	JointActionID     uint `json:"joint_action_id"`
	RomID             uint `json:"rom_id"`
	ClinicalFeatureID uint `json:"clinical_feature_id"`
	DegreeID          uint `json:"degree_id"`
}

type AfcResponse struct {
	BodyCompositionID uint    `json:"body_composition_id"`
	RomAv             *uint   `json:"rom_av"`
	ClinicalFeatureAv *string `json:"clinical_feature_av"`
	DegreeAv          *uint   `json:"degree_av"`
	RomName           *string `json:"rom_name"`
}

type GetAfcResponse struct {
	CreatedAdmin    string `json:"created_admin"`
	Created         string `json:"created"`
	GroupId         uint   `json:"group_id"`
	UserAfcResponse []UserAfcResponse
}

type UserAfcResponse struct {
	UpdatedAdmin      string `json:"updated_admin"`
	Updated           string `json:"updated"`
	BodyCompositionID uint   `json:"body_composition_id"`
	JointActionID     uint   `json:"joint_action_id"`
	RomID             *uint  `json:"rom_id"`
	ClinicalFeatureID *uint  `json:"clinical_feature_id"`
	DegreeID          *uint  `json:"degree_id"`
}

type SaveAfcRequest struct {
	Id   uint         `json:"-"`
	Uid  uint         `json:"uid"`
	Afcs []AfcRequest `json:"afcs"`
}

type SaveAfcHistoryRequest struct {
	Id      uint         `json:"-"`
	GroupId uint         `json:"group_id"`
	Afcs    []AfcRequest `json:"afcs"`
}

type AfcHistoryResponse struct {
	Id      uint         `json:"-"`
	GroupId uint         `json:"group_id"`
	Afcs    []AfcRequest `json:"afcs"`
}

type AgAdResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
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
