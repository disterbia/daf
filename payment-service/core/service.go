package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"payment-service/model"
	"time"

	"gorm.io/gorm"
)

type PaymentService interface {
	saveCart(uid, productOptionId uint) (string, error)
	getCart(uid uint) ([]ProductResponse, error)
	countCart(request CountRequest) (string, error)
	deleteCarts(uid uint, productOptionIds []uint) (string, error)

	paymentCallback(request PaymentCallbackResponse) (string, error)
	refund() (string, error)
}

type paymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) PaymentService {
	return &paymentService{db: db}
}

func (s *paymentService) deleteCarts(uid uint, productOptionIds []uint) (string, error) {
	// 장바구니에서 해당 상품 옵션 삭제
	if err := s.db.Where("uid = ? AND product_option_id IN ?", uid, productOptionIds).
		Delete(&model.Cart{}).Error; err != nil {
		return "", errors.New("db error")
	}

	return "1", nil
}

func (s *paymentService) countCart(request CountRequest) (string, error) {
	var cart model.Cart

	// 장바구니에서 해당 상품 옵션을 찾기
	if err := s.db.Where("uid = ? AND product_option_id = ?", request.Uid, request.ProductOptionId).
		First(&cart).Error; err != nil {
		return "", errors.New("cart item not found")
	}

	// 수량 증가 또는 감소
	if request.IsUp {
		cart.Quantity += 1
	} else {
		if cart.Quantity > 1 {
			cart.Quantity -= 1
		}
	}

	// 변경 사항 저장
	if err := s.db.Save(&cart).Error; err != nil {
		return "", errors.New("failed to update cart item")
	}

	return "1", nil
}

func (s *paymentService) saveCart(uid, productOptionId uint) (string, error) {

	// 같은 옵션이 장바구니에 있는지 확인
	var cart model.Cart
	if err := s.db.Where("uid = ? AND product_option_id = ?", uid, productOptionId).First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cart = model.Cart{
				Uid:             uid,
				ProductOptionID: productOptionId,
				Quantity:        1,
			}
			if err := s.db.Create(&cart).Error; err != nil {
				return "", errors.New("failed to add cart item")
			}
		} else {
			return "", errors.New("db error")
		}
	} else {
		cart.Quantity += 1
		if err := s.db.Save(&cart).Error; err != nil {
			return "", errors.New("failed to update cart item")
		}
	}

	return "1", nil
}
func (s *paymentService) getCart(uid uint) ([]ProductResponse, error) {
	var carts []model.Cart

	// 한 번의 쿼리로 모든 데이터 로드
	if err := s.db.
		Preload("ProductOption.Product").
		Where("uid = ?", uid).
		Find(&carts).Error; err != nil {
		return nil, errors.New("cart not found")
	}

	// 상품별로 그룹화
	productMap := make(map[uint]*ProductResponse)

	for _, item := range carts {
		product := item.ProductOption.Product
		option := item.ProductOption

		// 상품이 없으면 추가
		if _, exists := productMap[product.ID]; !exists {
			productMap[product.ID] = &ProductResponse{
				ID:        product.ID,
				Name:      product.Name,
				Price:     product.Price,
				SellPrice: product.SellPrice,
				Options:   []ProductOptionResponse{},
			}
		}

		// 옵션 추가
		productMap[product.ID].Options = append(productMap[product.ID].Options, ProductOptionResponse{
			OptionID:   option.ID,
			OptionName: option.Name,
			Price:      option.Price,
			Quantity:   item.Quantity,
		})
	}

	// 리스트로 변환
	var result []ProductResponse
	for _, product := range productMap {
		result = append(result, *product)
	}

	return result, nil
}
func (s *paymentService) paymentCallback(request PaymentCallbackResponse) (string, error) {
	log.Printf("결제 콜백 데이터: %+v\n", request)

	// 결제 성공 여부 확인
	if request.ResultCode != "0000" {
		return "", fmt.Errorf("결제 실패: %s", request.ResultMsg)
	}

	// 승인 요청 보내기
	signKey := "SU5JTElURV9UUklQTEVERVNfS0VZU1RS" // ⚠️ 실제 SIGN KEY 입력
	approvalResponse, err := sendApprovalRequest(request, signKey)
	if err != nil {
		return "", fmt.Errorf("승인 요청 실패: %v", err)
	}

	log.Printf("승인 응답 데이터: %+v\n", approvalResponse)

	// 승인 성공 여부 확인
	if approvalResponse.ResultCode != "0000" {
		return "", fmt.Errorf("승인 실패: %s", approvalResponse.ResultMsg)
	}

	return approvalResponse.Tid, nil
}

func (s *paymentService) refund() (string, error) {
	// 환경 변수에서 설정값 가져오기
	mid := "INIpayTest"                               // ⚠️ 실제 상점 아이디 입력
	iniApiKey := "ItEQKi3rY7uvDS8l"                   // ⚠️ 이니시스에서 제공하는 API Key
	clientIp := "192.168.1.1"                         // ⚠️ 실제 서버 IP 입력
	tid := "StdpayCARDINIpayTest20250312150916921936" // ⚠️ 취소할 승인 TID (AuthToken 사용)
	reason := "고객 요청에 의한 결제 취소"                       // ⚠️ 취소 사유

	// 현재 타임스탬프 생성 (YYYYMMDDhhmmss)
	timestamp := time.Now().Format("20060102150405")

	// `data` JSON 문자열 생성
	dataMap := map[string]string{
		"tid": tid,
		"msg": reason,
	}
	dataJSON, _ := json.Marshal(dataMap) // JSON 직렬화
	// ⚠️ **hashData 형식 맞추기**
	plainText := fmt.Sprintf("%s%s%s%s%s", iniApiKey, mid, "refund", timestamp, string(dataJSON))
	hashData := generateSHA512Hash(plainText)

	// 최종 요청 데이터 구성
	refundReq := map[string]interface{}{
		"mid":       mid,
		"type":      "refund",
		"timestamp": timestamp,
		"clientIp":  clientIp,
		"hashData":  hashData,
		"data":      dataMap, // ⚠️ `data`를 JSON이 아닌 Object로 전달 (문서 기준)
	}

	// JSON 변환
	requestBody, _ := json.Marshal(refundReq)

	// HTTP POST 요청
	url := "https://iniapi.inicis.com/v2/pg/refund"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("취소 요청 생성 실패: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("취소 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	// 응답 데이터 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("응답 데이터 읽기 실패: %v", err)
	}

	// JSON 응답 데이터 파싱
	refundResp := &RefundResponse{}
	err = json.Unmarshal(body, refundResp)
	if err != nil {
		return "", fmt.Errorf("응답 JSON 파싱 실패: %v", err)
	}

	// 취소 결과 출력
	log.Printf("취소 응답 데이터: %+v\n", refundResp)

	// 취소 성공 여부 확인
	if refundResp.ResultCode == "00" {
		return fmt.Sprintf("결제 취소 성공! 취소일자: %s, 취소시간: %s", refundResp.CancelDate, refundResp.CancelTime), nil
	}

	return "", fmt.Errorf("취소 실패: %s (코드: %s)", refundResp.ResultMsg, refundResp.DetailResultCode)
}
