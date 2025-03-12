package core

type LoginResponse struct {
	Jwt      string `json:"jwt,omitempty"`
	SnsId    string `json:"sns_id,omitempty"`
	SnsEmail string `json:"sns_email,omitempty"`
}

type VerifyRequest struct {
	PhoneNumber string `json:"phone_number" example:"01000000000"`
	Code        string `json:"code" example:"인증번호 6자리"`
}

// Authorization Code를 담는 요청 구조체
type CallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// 애플 서버의 응답을 매핑하는 구조체
type AppleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Code         string `json:"code"`
}

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type KakaoTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type FacebookTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type FacebookUserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type NaverTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type NaverResponse struct {
	Response struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"response"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type FindUsernameRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone" example:"01000000000"`
}

type FindPasswordRequest struct {
	Username string `json:"username"`
	Phone    string `json:"phone" example:"01000000000"`
}

type SignInRequest struct {
	SnsId        string `json:"sns_id"`
	SnsEmail     string `json:"sns_email"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	Gender       bool   `json:"gender"`
	Birth        string `json:"birth" example:"yyyy-mm-dd"`
	Phone        string `json:"phone"`
	Addr         string `json:"addr"`
	AddrDetail   string `json:"addr_detail"`
	IsAgree      bool   `json:"is_agree"`
	DisableType  uint   `json:"disable_type"`
	VisitPurpose uint   `json:"visit_purpose"`
}

type SetUserRequest struct {
	Uid          uint   `json:"-"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Addr         string `json:"addr"`
	AddrDetail   string `json:"addr_detail"`
	IsAgree      bool   `json:"is_agree"`
	DisableType  uint   `json:"disable_type"`
	VisitPurpose uint   `json:"visit_purpose"`
}

type AutoLoginRequest struct {
	Email    string `json:"-"`
	FcmToken string `json:"fcm_token"`
	DeviceId string `json:"device_id"`
}

type UserResponse struct {
	Username     string `json:"username"`
	Name         string `json:"name"`
	Gender       bool   `json:"gender"`
	Birth        string `json:"birth"`
	Phone        string `json:"phone"`
	Addr         string `json:"addr"`
	AddrDetail   string `json:"addr_detail"`
	IsAgree      bool   `json:"is_agree"`
	DisableType  uint   `json:"disable_type"`
	VisitPurpose uint   `json:"visit_purpose"`
}

type UserRequest struct {
	ID           uint   `json:"-"`
	Nickname     string `json:"nickname"`
	ProfileImage string `json:"profile_image" example:"base64string"`
}

type ImageResponse struct {
	Url          string `json:"url"`
	ThumbnailUrl string `json:"thumbnail_url"`
}

type AppVersionResponse struct {
	LatestVersion string `json:"latest_version"`
	AndroidLink   string `json:"android_link"`
	IosLink       string `json:"ios_link"`
}

type BasicResponse struct {
	Code string `json:"code"`
}

// // for swagger ////
type SuccessResponse struct {
	Jwt string `json:"jwt"`
}
type ErrorResponse struct {
	Err string `json:"error"`
}

type PaymentCallbackResponse struct {
	ResultCode   string `json:"resultCode"`
	ResultMsg    string `json:"resultMsg"`
	Mid          string `json:"mid"`
	OrderNumber  string `json:"orderNumber"`
	AuthToken    string `json:"authToken"`
	IdcName      string `json:"idc_name"`
	AuthUrl      string `json:"authUrl"`
	NetCancelUrl string `json:"netCancelUrl"`
	Charset      string `json:"charset"`
	MerchantData string `json:"merchantData"`
}

type PaymentApprovalResponse struct {
	ResultCode string `json:"resultCode"`
	ResultMsg  string `json:"resultMsg"`
	Tid        string `json:"tid"`
	Mid        string `json:"mid"`
	MOID       string `json:"MOID"`
	TotPrice   string `json:"TotPrice"`
	GoodName   string `json:"goodName"`
	PayMethod  string `json:"payMethod"`
	ApplDate   string `json:"applDate"`
	ApplTime   string `json:"applTime"`
	EventCode  string `json:"EventCode"`
	BuyerName  string `json:"buyerName"`
	BuyerTel   string `json:"buyerTel"`
	BuyerEmail string `json:"buyerEmail"`
	CustEmail  string `json:"custEmail"`
}

// 🔹 이니시스 결제 취소 요청 구조체
type RefundRequest struct {
	Mid       string `json:"mid"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	ClientIp  string `json:"clientIp"`
	HashData  string `json:"hashData"`
	Data      struct {
		Tid string `json:"tid"`
		Msg string `json:"msg"`
	} `json:"data"`
}

// 🔹 이니시스 결제 취소 응답 구조체
type RefundResponse struct {
	ResultCode       string `json:"resultCode"`
	ResultMsg        string `json:"resultMsg"`
	CancelDate       string `json:"cancelDate"`
	CancelTime       string `json:"cancelTime"`
	CshrCancelNum    string `json:"cshrCancelNum"`
	DetailResultCode string `json:"detailResultCode"`
	ReceiptInfo      string `json:"receiptInfo"`
}
