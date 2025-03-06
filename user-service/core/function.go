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
	"net/url"
	"os"
	"strings"
	"time"
	"user-service/model"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

func appleLogin(idToken string) (string, error) {
	jwks, err := getApplePublicKeys()
	if err != nil {
		return "", err
	}

	parsedToken, err := verifyAppleIDToken(idToken, jwks)
	if err != nil {
		return "", err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		sub, ok := claims["sub"].(string)
		if !ok {
			return "", errors.New("email not found in token claims")
		}

		return sub, nil

	}
	return "", errors.New("invalid token")

}
func kakaoLogin(idToken string) (string, error) {
	jwks, err := getKakaoPublicKeys()
	if err != nil {
		return "", err
	}

	parsedToken, err := verifyKakaoTokenSignature(idToken, jwks)
	if err != nil {
		return "", err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		sub, ok := claims["sub"].(string)
		if !ok {
			return "", errors.New("email not found in token claims")
		}

		return sub, nil
	}
	return "", errors.New("invalid token")

}

func googleLogin(idToken, clientID string) (string, error) {
	sub, err := validateGoogleIDToken(idToken, clientID)
	if err != nil {
		return "", err
	}
	return sub, nil
}

// login 함수: 사용자가 없으면 이메일을 키로 Redis에 저장 (10분 후 삭제)
func snsLogin(snsId string, snsType uint, service *userService) (LoginResponse, error) {
	var user model.User
	if err := service.db.Where("sns_id = ?", snsId).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		// 사용자가 없으면 Redis에 저장 (10분 후 자동 삭제)
		ctx := context.Background()
		key := snsId                        // 이메일 자체를 키로 사용
		value := fmt.Sprintf("%d", snsType) // snsType을 값으로 저장

		if err := service.redisClient.Set(ctx, key, value, 10*time.Minute).Err(); err != nil {
			log.Println(err)
			return LoginResponse{}, errors.New("fail to login")
		}
		return LoginResponse{SnsId: snsId}, nil
	} else if err != nil {
		return LoginResponse{}, errors.New("db error")
	}
	// JWT 토큰 생성
	tokenString, err := generateJWT(user)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{Jwt: tokenString}, nil
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

	// sub 추출
	sub, ok := payload.Claims["sub"].(string)
	if !ok {
		return "", errors.New("email claim not found in token")
	}

	return sub, nil
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
	if userInfo.Response.ID == "" {
		return "", fmt.Errorf("email not found in Naver user info")
	}
	return userInfo.Response.ID, nil
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

	if userResponse.ID == "" {
		return "", errors.New("email not found in Facebook account")
	}

	return userResponse.ID, nil
}

func sendCode(number, code string) error {

	apiURL := "https://apis.aligo.in/send/"
	data := url.Values{}
	data.Set("key", os.Getenv("API_KEY"))
	data.Set("user_id", os.Getenv("USER_ID"))
	data.Set("sender", os.Getenv("SENDER"))
	data.Set("receiver", number)
	data.Set("msg", "인증번호는 ["+code+"]"+" 입니다.")

	// HTTP POST 요청 실행
	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		fmt.Printf("HTTP Request Failed: %s\n", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println(fmt.Errorf("server returned non-200 status: %d, body: %s", resp.StatusCode, string(body)))

	return nil

}
