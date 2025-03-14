package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"payment-service/model"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentService interface {
	saveCart(uid, productOptionId uint) (string, error)
	getCart(uid uint) ([]ProductResponse, error)
	countCart(req CountRequest) (string, error)
	deleteCarts(req DeleteCartRequest) (string, error)
	getSales(req GetSalesRequest) (SaleResponse, error)

	payment(req PaymentRequest) (string, error)

	paymentCallback(req PaymentCallbackResponse) (string, error)
	refund() (string, error)
}

type paymentService struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewPaymentService(db *gorm.DB, redisClient *redis.Client) PaymentService {
	return &paymentService{db: db, redisClient: redisClient}
}

func (s *paymentService) getSales(req GetSalesRequest) (SaleResponse, error) {

	var totalPoint uint
	if err := s.db.Model(&model.UserPoint{}).
		Where("uid = ?", req.Uid).
		Select("COALESCE(SUM(point), 0)").
		Scan(&totalPoint).Error; err != nil {
		return SaleResponse{}, errors.New("db error")
	}

	var userCoupons []model.UserCoupon
	// 해당 유저가 보유한 쿠폰 리스트 조회
	if err := s.db.Where("uid = ?", req.Uid).Preload("Coupon").Find(&userCoupons).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SaleResponse{Point: totalPoint}, nil
		}
		return SaleResponse{}, errors.New("db error 2")
	}

	currentDate := time.Now().Truncate(24 * time.Hour)

	// 유효 기간 체크
	var possibleCoupons []model.Coupon
	for _, uc := range userCoupons {
		coupon := uc.Coupon
		if coupon.DueDate.Before(currentDate) {
			continue // 만료된 쿠폰은 제외
		}
		possibleCoupons = append(possibleCoupons, coupon)
	}

	//  유효한 쿠폰 필터링
	var validCoupons []CouponResponse
	for _, coupon := range possibleCoupons {
		// 특정 상품에만 적용 가능한 경우 필터링
		if len(coupon.PossibleProductIds) > 0 {
			isApplicable := false
			for _, productId := range req.ProductIds { // 유저가 구매하려는 상품 ID 목록 반복
				for _, possibleId := range coupon.PossibleProductIds { // 쿠폰이 적용 가능한 상품 ID 목록 반복
					if productId == possibleId {
						isApplicable = true
						break //  현재 상품이 쿠폰 적용 가능 목록에 포함되면 내부 루프 종료
					}
				}
				if isApplicable {
					break //  하나의 상품이라도 적용 가능하면 외부 루프 종료
				}
			}
			if !isApplicable {
				continue // 유저의 상품 목록 중 어느 것도 해당 쿠폰을 사용할 수 없다면 제외
			}
		}

		// 4️⃣ 쿠폰을 응답 리스트에 추가
		validCoupons = append(validCoupons, CouponResponse{
			Id:        coupon.ID,
			Name:      coupon.Name,
			Detail:    coupon.Detail,
			Price:     coupon.Price,
			Percent:   coupon.Percent,
			DueDate:   coupon.DueDate,
			CanDouble: coupon.CanDouble,
		})
	}

	return SaleResponse{Point: totalPoint, Coupons: validCoupons}, nil
}

func (s *paymentService) payment(request PaymentRequest) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 함수 종료 시 컨텍스트 해제

	price, err := calculatePrice(s.db, request)
	if err != nil {
		return "", err
	}
	request.Price = price
	// JSON으로 변환
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("JSON 직렬화 실패: %v", err)
	}

	oid := uuid.New().String()
	err = s.redisClient.Set(ctx, oid, jsonData, 10*time.Minute).Err()
	if err != nil {
		return "", errors.New("internal error")
	}

	return oid, nil
}

func (s *paymentService) deleteCarts(req DeleteCartRequest) (string, error) {
	// 장바구니에서 해당 상품 옵션 삭제
	if err := s.db.Where("uid = ? AND product_option_id IN ?", req.Uid, req.ProductOptionIds).
		Delete(&model.Cart{}).Error; err != nil {
		return "", errors.New("db error")
	}

	return "1", nil
}

func (s *paymentService) countCart(req CountRequest) (string, error) {
	var cart model.Cart

	// 장바구니에서 해당 상품 옵션을 찾기
	if err := s.db.Where("uid = ? AND product_option_id = ?", req.Uid, req.ProductOptionId).
		First(&cart).Error; err != nil {
		return "", errors.New("cart item not found")
	}

	// 수량 증가 또는 감소
	if req.IsUp {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 함수 종료 시 컨텍스트 해제

	jsonData, err := s.redisClient.Get(ctx, request.OrderNumber).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("결제 요청이 유효하지 않음")
	} else if err != nil {
		return "", fmt.Errorf("internal error")
	}

	var data PaymentRequest
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return "", fmt.Errorf("JSON 파싱 실패: %v", err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		} else if tx.Error != nil {
			tx.Rollback()
		}
	}()

	// db 저장 CouponHistory

	// 승인 요청 보내기
	signKey := "SU5JTElURV9UUklQTEVERVNfS0VZU1RS" // ⚠️ 실제 SIGN KEY 입력
	approvalResponse, err := sendApprovalRequest(request, signKey, data.Price)
	if err != nil {
		return "", fmt.Errorf("승인 요청 실패: %v", err)
	}

	log.Printf("승인 응답 데이터: %+v\n", approvalResponse)

	// 승인 성공 여부 확인
	if approvalResponse.ResultCode != "0000" {
		return "", fmt.Errorf("승인 실패: %s", approvalResponse.ResultMsg)
	}

	if err = s.redisClient.Del(ctx, request.OrderNumber).Err(); err != nil {
		log.Println(err)
		return "", errors.New("internal error2")
	}

	//tid 저장

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
