package core

import (
	"admin_service/model"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtSecretKey = []byte("adapfit_mark")

func verifyJWT(c *gin.Context) (uint, string, error) {
	// 헤더에서 JWT 토큰 추출
	tokenString := c.GetHeader("Authorization")
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

func generateJWT(admin model.Admin) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    admin.ID,
		"email": admin.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(), // 한달 유효 기간
	})

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func copyStruct(input interface{}, output interface{}) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, output)
	if err != nil {
		return err
	}

	return nil
}
func validateEmail(email string) error {
	// 빈 문자열 검사
	if email == "" || len(email) > 50 || strings.Contains(email, " ") {
		return errors.New("invalid email format")
	}

	// 이메일 검증을 위한 정규 표현식
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}
