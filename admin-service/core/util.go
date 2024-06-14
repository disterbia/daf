package core

import (
	"admin-service/model"
	"errors"
	"reflect"
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

func copyStruct(src, dst interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldName := srcVal.Type().Field(i).Name

		dstField := dstVal.FieldByName(srcFieldName)
		if dstField.IsValid() && dstField.Type() == srcField.Type() {
			dstField.Set(srcField)
		}
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

func validateSignIn(request SignInRequest) error {
	pattern := `^010\d{8}$`
	matched, err := regexp.MatchString(pattern, request.Phone)
	if err != nil || !matched {
		return errors.New("invalid phone format, should be 01000000000")
	}

	name := strings.TrimSpace(request.Name)
	if len(name) > 10 || len(name) == 0 {
		return errors.New("invalid name")
	}
	return nil
}

func validateSaveUser(request SaveUserRequest) error {
	pattern := `^010\d{8}$`
	matched, err := regexp.MatchString(pattern, request.Phone)
	if err != nil || !matched {
		return errors.New("invalid phone format, should be 01000000000")
	}

	name := strings.TrimSpace(request.Name)
	if len(name) > 10 || len(name) == 0 {
		return errors.New("invalid name")
	}
	return nil
}

func contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func checkDuplicates(slice []uint) bool {
	seen := make(map[uint]struct{})
	for _, item := range slice {
		if _, exists := seen[item]; exists {
			return true // Duplicate found
		}
		seen[item] = struct{}{}
	}
	return false // No duplicates
}
