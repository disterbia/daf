package core

import (
	"errors"
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

func intersect(slice1, slice2 []uint) []uint {
	map1 := make(map[uint]bool)
	for _, v := range slice1 {
		map1[v] = true
	}

	var intersection []uint
	for _, v := range slice2 {
		if _, found := map1[v]; found {
			intersection = append(intersection, v)
		}
	}

	return intersection
}

func mergeAndRemoveDuplicates(slice1, slice2 []uint) []uint {
	elementMap := make(map[uint]bool)
	var result []uint

	// 슬라이스1의 요소를 맵에 추가
	for _, v := range slice1 {
		if _, exists := elementMap[v]; !exists {
			elementMap[v] = true
			result = append(result, v)
		}
	}

	// 슬라이스2의 요소를 맵에 추가
	for _, v := range slice2 {
		if _, exists := elementMap[v]; !exists {
			elementMap[v] = true
			result = append(result, v)
		}
	}

	return result
}

func recommend() {}
