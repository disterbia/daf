package core

import (
	"coach-service/model"
	"encoding/json"
	"errors"
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

func validateRecommendRequest(request RecommendRequest) error {
	if request.ExerciseID == 0 || request.MachineIDs == nil || request.PurposeIDs == nil || request.BodyRomClinicDegree == nil ||
		len(request.MachineIDs) == 0 || len(request.PurposeIDs) == 0 || len(request.BodyRomClinicDegree) == 0 {
		return errors.New("check body")
	}
	return nil
}
func getChosung(input string) string {
	var result []rune
	for _, r := range input {
		if r >= 0xAC00 && r <= 0xD7A3 {
			// 음절을 초성으로 변환
			initial := (r - 0xAC00) / 588
			initialChar := rune(0x1100 + initial)
			result = append(result, initialChar)
		} else if r >= 0x3131 && r <= 0x318E {
			// 자모 문자 처리
			result = append(result, r)
		} else {
			// 기타 문자 처리
			result = append(result, r)
		}
	}
	return string(result)
}

func uintPointer(u uint) *uint {
	return &u
}
