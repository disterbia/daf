package core

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

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

func validateRequest(request UserJointActionRequest) error {
	v := reflect.ValueOf(request)
	t := v.Type()

	// 정규 표현식으로 각 필드 유효성 검사
	pattern := `(?i)^[1-5][ntpcswa][1-7]$`
	re := regexp.MustCompile(pattern)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// uint 타입 필드 또는 "-" 태그가 있는 필드는 건너뜁니다.
		if fieldType.Type.Kind() == reflect.Uint || fieldType.Tag.Get("-") == "-" {
			continue
		}

		// string 타입 필드에 대해서만 검증
		if field.Type().Kind() == reflect.String {
			str := field.String()
			if str == "" {
				continue
			}
			// 정규 표현식에 맞지 않는 경우
			if !re.MatchString(str) {
				errorMsg := fmt.Sprintf("Invalid field %s: %s", fieldType.Name, str)
				return errors.New(errorMsg)
			}
		}
	}

	return nil
}
