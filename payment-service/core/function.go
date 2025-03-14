package core

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SHA256 해시 생성 함수
func generateSHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// 승인 요청 함수
func sendApprovalRequest(request PaymentCallbackResponse, signKey string, price uint) (*PaymentApprovalResponse, error) {
	log.Printf("결제 콜백 데이터: %+v\n", request)
	log.Printf("받은 IDC센터 코드: %s\n", request.IdcName)
	// IDC센터 코드에 따른 승인 URL 매핑
	idcUrls := map[string]string{
		"fc":  "https://fcstdpay.inicis.com/api/payAuth",
		"ks":  "https://ksstdpay.inicis.com/api/payAuth",
		"stg": "https://stgstdpay.inicis.com/api/payAuth",
	}

	// `idc_name`이 비어있다면 `authUrl` 기반으로 자동 설정
	if request.IdcName == "" {
		request.IdcName = detectIDCName(request.AuthUrl)
		log.Printf("자동 감지된 IDC센터 코드: %s\n", request.IdcName)
	}

	// `idc_name`이 올바른지 검증
	expectedAuthUrl, validIDC := idcUrls[request.IdcName]
	if !validIDC {
		return nil, fmt.Errorf("알 수 없는 IDC센터 코드: %s", request.IdcName)
	}

	// `authUrl`이 IDC센터의 승인 URL과 일치하는지 검증
	if request.AuthUrl != expectedAuthUrl {
		return nil, fmt.Errorf("승인 요청 URL이 IDC센터 코드와 일치하지 않음. 예상 URL: %s, 받은 URL: %s", expectedAuthUrl, request.AuthUrl)
	}
	// 현재 타임스탬프 생성
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// SHA256 해시값 생성
	signature := generateSHA256Hash(fmt.Sprintf("authToken=%s&timestamp=%s", request.AuthToken, timestamp))
	verification := generateSHA256Hash(fmt.Sprintf("authToken=%s&signKey=%s&timestamp=%s", request.AuthToken, signKey, timestamp))

	// 승인 요청 데이터 설정 (application/x-www-form-urlencoded)
	formData := url.Values{}
	formData.Set("mid", request.Mid)
	formData.Set("authToken", request.AuthToken)
	formData.Set("timestamp", timestamp)
	formData.Set("signature", signature)
	formData.Set("verification", verification)
	formData.Set("charset", "UTF-8")
	formData.Set("format", "JSON") // JSON 응답을 요청
	formData.Set("price", strconv.FormatUint(uint64(price), 10))
	// 승인 요청 (HTTP POST)
	resp, err := http.Post(request.AuthUrl, "application/x-www-form-urlencoded", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("승인 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	// 응답 데이터 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 데이터 읽기 실패: %v", err)
	}

	// JSON 응답 데이터 파싱
	approvalResponse := &PaymentApprovalResponse{}
	err = json.Unmarshal(body, approvalResponse)
	if err != nil {
		return nil, fmt.Errorf("응답 JSON 파싱 실패: %v", err)
	}

	return approvalResponse, nil
}

// 🔹 SHA512 해시 생성 함수 (취소 요청)
func generateSHA512Hash(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

func detectIDCName(authUrl string) string {
	if strings.Contains(authUrl, "fcstdpay.inicis.com") {
		return "fc"
	} else if strings.Contains(authUrl, "ksstdpay.inicis.com") {
		return "ks"
	} else if strings.Contains(authUrl, "stgstdpay.inicis.com") {
		return "stg"
	}
	return "" //  알 수 없는 경우
}

func calculatePrice(db *gorm.DB, request PaymentRequest) (uint, error) {
	return 100, nil
}
