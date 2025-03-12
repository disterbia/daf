package core

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"
	"user-service/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var jwtSecretKey = []byte("haruharu_mark_user")

type PublicKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
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
		"exp":   time.Now().Add(time.Hour * 24 * 1).Unix(), // 하루 유효 기간
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

func validateDate(dateStr string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, errors.New("invalid date format, should be YYYY-MM-DD")
	}
	return date, nil
}

func validateUsername(username string) error {
	// 사용자명 검증 (4~20자, 특수문자 포함 불가)
	value := strings.TrimSpace(username)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9]{4,20}$`)
	if !usernameRegex.MatchString(value) {
		return errors.New("invalid username format (4~20 characters, no special characters)")
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

func validateSignIn(request SignInRequest, snsId *string) error {

	// 전화번호 검증 (010으로 시작하는 11자리 숫자)
	phoneRegex := regexp.MustCompile(`^010\d{8}$`)
	if !phoneRegex.MatchString(request.Phone) {
		return errors.New("invalid phone format, should be 010xxxxxxxx")
	}
	if snsId != nil {
		// 아이디 검증 (4~20자, 특수문자 포함 불가)
		username := strings.TrimSpace(request.Username)
		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9]{4,20}$`)
		if !usernameRegex.MatchString(username) {
			return errors.New("invalid username format (4~20 characters, no special characters)")
		}

		// 비밀번호 검증 (최소 8~20자, 영문 대소문자/숫자/특수문자 중 2종류 이상 포함)
		if !checkPassword(request.Password) {
			return errors.New("invalid password format (must include at least two of: letters, numbers, special characters, and be at least 8 characters long)")
		}
	}

	// 사용자명 검증 (4~20자, 특수문자 포함 불가)
	name := strings.TrimSpace(request.Name)
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9]{4,20}$`)
	if !nameRegex.MatchString(name) {
		return errors.New("invalid name format (4~20 characters, no special characters)")
	}

	return nil
}

func checkPassword(value string) bool {
	password := strings.TrimSpace(value)
	// 길이 검사: 8~20자
	if len(password) < 8 || len(password) > 20 {
		return false
	}

	// 각 카테고리 포함 여부 확인
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 카테고리 중 2개 이상 포함 여부 확인
	count := 0
	if hasLower {
		count++
	}
	if hasUpper {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 2
}

// 🔥 회원 고유 코드 생성 (이메일 또는 휴대폰 기반 UUID v5)
func generateMemberCode(identifier string) string {
	// 네임스페이스 UUID (고정값, 동일한 네임스페이스에서 생성해야 동일한 값 유지)
	namespaceUUID := uuid.NameSpaceDNS

	// UUID v5 생성 (SHA-1 기반)
	newUUID := uuid.NewSHA1(namespaceUUID, []byte(identifier))

	// 앞 3바이트(6자리)만 추출하여 반환
	hash := sha1.Sum(newUUID[:])
	code := hex.EncodeToString(hash[:])[:6]

	return strings.ToUpper(code) // 대문자로 변환
}
