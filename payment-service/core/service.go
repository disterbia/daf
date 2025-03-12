package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type PaymentService interface {
	paymentCallback(request PaymentCallbackResponse) (string, error)
	refund() (string, error)
}

type paymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) PaymentService {
	return &paymentService{db: db}
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
