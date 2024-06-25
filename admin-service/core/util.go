package core

import (
	"admin-service/model"
	"errors"
	"fmt"
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

func calculateAgeCode(birthday time.Time) uint {
	now := time.Now()
	age := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		age--
	}

	switch {
	case age < 10:
		return 1
	case age < 20:
		return 2
	case age < 30:
		return 3
	case age < 40:
		return 4
	case age < 50:
		return 5
	case age < 60:
		return 6
	default:
		return 7
	}
}

func getBirthdayRangeByAgeCode(ageCode uint) (time.Time, time.Time, error) {
	now := time.Now()
	currentYear := now.Year()
	var startYear, endYear int

	switch ageCode {
	case 1: // 0-9 years
		startYear = currentYear - 9
		endYear = currentYear
	case 2: // 10-19 years
		startYear = currentYear - 19
		endYear = currentYear - 10
	case 3: // 20-29 years
		startYear = currentYear - 29
		endYear = currentYear - 20
	case 4: // 30-39 years
		startYear = currentYear - 39
		endYear = currentYear - 30
	case 5: // 40-49 years
		startYear = currentYear - 49
		endYear = currentYear - 40
	case 6: // 50-59 years
		startYear = currentYear - 59
		endYear = currentYear - 50
	case 7: // 60 years and above
		startYear = 0
		endYear = currentYear - 60
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid age code")
	}

	startDate := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(endYear, 12, 31, 23, 59, 59, 999999999, time.UTC)
	return startDate, endDate, nil
}

func validateAfc(request []AfcRequest) bool {
	for _, v := range request {
		if v.BodyCompositionID == 0 || v.JointActionID == 0 {
			return false
		} else if v.BodyCompositionID != uint(TR) && v.BodyCompositionID != uint(LOCOMOTION) && v.ClinicalFeatureID != uint(AC) &&
			(v.RomID == 0 || v.ClinicalFeatureID == 0 || v.DegreeID == 0) {
			return false
		}
	}
	return true
}
func sum(slice []uint) uint {
	total := uint(0)
	for _, value := range slice {
		total += value
	}
	return total
}
