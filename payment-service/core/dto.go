package core

type ProductResponse struct {
	ID        uint                    `json:"id"`
	Name      string                  `json:"name"`
	Price     uint                    `json:"price"`
	SellPrice uint                    `json:"sell_price"`
	Options   []ProductOptionResponse `json:"options"`
}

type ProductOptionResponse struct {
	OptionID   uint   `json:"optionId"`
	OptionName string `json:"optionName"`
	Price      uint   `json:"price"`
	Quantity   uint   `json:"quantity"`
}

type CountRequest struct {
	Uid             uint `json:"-"`
	ProductOptionId uint `json:"product_option_id"`
	IsUp            bool `json:"is_up"`
}
type DeleteCartRequest struct {
	Uid             uint `json:"-"`
	ProductId       uint `json:"product_id"`
	ProductOptionId uint `json:"product_option_id"`
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

type BasicResponse struct {
	Code string `json:"code"`
}
