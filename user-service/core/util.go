package core

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"
	"user-service/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/idtoken"
)

var jwtSecretKey = []byte("haruharu_mark_user")

type PublicKey struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []PublicKey `json:"keys"`
}

func verifyJWT(c *fiber.Ctx) (uint, string, error) {
	// 헤더에서 JWT 토큰 추출
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return 0, "", errors.New("authorization header is required")
	}

	// 'Bearer ' 접두사 제거
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if err != nil || !token.Valid {
		return 0, "", errors.New("invalid token")
	}

	id := uint((*claims)["id"].(float64))
	email := (*claims)["email"].(string)
	if email == "" || id == 0 {
		return 0, "", errors.New("id or email not found in token")
	}
	return id, email, nil
}

func generateJWT(user model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(), // 한달 유효 기간
	})

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func decodeJwt(tokenString string) string {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Println(err)
	}

	// MapClaims 타입으로 claims 확인
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// 'iss' 확인
		if iss, ok := claims["iss"].(string); ok {
			fmt.Println("Issuer (iss):", iss)
			return iss
		} else {
			fmt.Println("'iss' 이 없습니다.")
			return ""
		}
	} else {
		log.Println("클레임을 MapClaims로 변환할 수 없습니다.")
		return ""
	}
}

// Apple의 client_secret을 생성하는 함수
func GenerateClientSecret(keyID, teamID, clientID, privateKey string) (string, error) {
	// JWT 클레임 설정
	claims := jwt.MapClaims{
		"iss": teamID,                               // Team ID
		"iat": time.Now().Unix(),                    // 현재 시간
		"exp": time.Now().Add(6 * time.Hour).Unix(), // 만료 시간 (최대 6개월)
		"aud": "https://appleid.apple.com",          // Audience
		"sub": clientID,                             // Service ID
	}

	// JWT 생성 및 헤더에 키 ID 추가
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	// PEM 포맷의 비공개 키 파싱
	parsedKey, err := parsePrivateKey(privateKey)
	if err != nil {
		log.Println("parse")
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// JWT 서명 생성
	return token.SignedString(parsedKey)
}

// PEM 형식의 개인 키를 파싱하는 함수 (PKCS8 지원)
func parsePrivateKey(privateKey string) (*ecdsa.PrivateKey, error) {
	// PEM 블록 추출
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Println("pem")
		return nil, errors.New("invalid private key: not a valid PEM block")
	}

	// PKCS8 형식의 키 파싱
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 키 타입 확인 및 변환
	ecPrivateKey, ok := parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		log.Println("ecdsa")
		return nil, errors.New("private key is not of type ECDSA")
	}

	return ecPrivateKey, nil
}

// Apple 공개키 가져오기
func getApplePublicKeys() (JWKS, error) {
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return JWKS{}, err
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return JWKS{}, err
	}
	return jwks, nil
}

// Apple 공개키로 서명 검증
func verifyAppleIDToken(token string, jwks JWKS) (*jwt.Token, error) {
	kid, err := extractKidFromToken(token)
	if err != nil {
		return nil, err
	}

	var key *rsa.PublicKey
	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
			if err != nil {
				return nil, err
			}
			eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
			if err != nil {
				return nil, err
			}

			n := big.NewInt(0).SetBytes(nBytes)
			e := big.NewInt(0).SetBytes(eBytes).Int64()
			key = &rsa.PublicKey{N: n, E: int(e)}
			break
		}
	}

	if key == nil {
		return nil, errors.New("appropriate public key not found")
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	return parsedToken, nil
}

// 카카오 공개키 가져오기
func getKakaoPublicKeys() (JWKS, error) {
	resp, err := http.Get("https://kauth.kakao.com/.well-known/jwks.json")
	if err != nil {
		return JWKS{}, err
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return JWKS{}, err
	}
	return jwks, nil
}

// 카카오 공개키로 서명 검증
func verifyKakaoTokenSignature(token string, jwks JWKS) (*jwt.Token, error) {
	kid, err := extractKidFromToken(token)
	if err != nil {
		return nil, err
	}

	var key *rsa.PublicKey
	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
			if err != nil {
				return nil, err
			}
			eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
			if err != nil {
				return nil, err
			}

			n := big.NewInt(0).SetBytes(nBytes)
			e := big.NewInt(0).SetBytes(eBytes).Int64()
			key = &rsa.PublicKey{N: n, E: int(e)}
			break
		}
	}

	if key == nil {
		return nil, errors.New("appropriate public key not found")
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	return parsedToken, nil
}

// ID 토큰에서 kid 추출
func extractKidFromToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid token format")
	}
	headerPart := parts[0]
	headerJson, err := base64.RawURLEncoding.DecodeString(headerPart)
	if err != nil {
		return "", err
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerJson, &header); err != nil {
		return "", err
	}

	kid, ok := header["kid"].(string)
	if !ok {
		return "", errors.New("kid not found in token header")
	}
	return kid, nil
}

// Google ID 토큰을 검증하고 이메일을 반환
func validateGoogleIDToken(idToken, clientID string) (string, error) {
	log.Print("idToken: ", idToken)
	// idtoken 패키지를 사용하여 토큰 검증
	payload, err := idtoken.Validate(context.Background(), idToken, clientID)
	if err != nil {
		log.Printf("Token validation error: %v", err)
		return "", err
	}

	// 이메일 추출
	email, ok := payload.Claims["email"].(string)
	if !ok {
		return "", errors.New("email claim not found in token")
	}

	return email, nil
}

func getFacebookUserInfo(accessToken string) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/me?fields=id,name,email&access_token=%s", accessToken)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get user info, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var userResponse FacebookUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return "", err
	}

	if userResponse.Email == "" {
		return "", errors.New("email not found in Facebook account")
	}

	return userResponse.Email, nil
}

func getNaverUserInfo(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://openapi.naver.com/v1/nid/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get user info, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var userInfo NaverResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", err
	}
	if userInfo.Response.Email == "" {
		return "", fmt.Errorf("email not found in Naver user info")
	}
	return userInfo.Response.Email, nil
}

func validateDate(dateStr string) error {
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.New("invalid date format, should be YYYY-MM-DD")
	}
	return nil
}

func validateTime(timeStr string) error {
	if len(timeStr) != 5 {
		return errors.New("invalid time format, should be HH:MM")
	}
	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		return errors.New("invalid time format, should be HH:MM")
	}
	return nil
}

func validatePhoneNumber(phone string) error {
	// 정규 표현식 패턴: 010으로 시작하며 총 11자리 숫자
	pattern := `^010\d{8}$`
	matched, err := regexp.MatchString(pattern, phone)
	if err != nil || !matched {
		return errors.New("invalid phone format, should be 01000000000")
	}
	return nil
}
