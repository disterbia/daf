package core

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	kitEndpoint "github.com/go-kit/kit/endpoint"
)

var userLocks sync.Map

// @Tags 회원 신체능력 /daf
// @Summary 회원 신체능력 설정
// @Description 회원 신체능력 설정시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body UserJointActionRequest true "요청 DTO - 업데이트 할 데이터"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /set-user-daf [post]
func SetUserHandler(endpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 토큰 검증 및 처리
		id, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 사용자별 잠금 시작
		if _, loaded := userLocks.LoadOrStore(id, true); loaded {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Concurrent request detected"})
			return
		}
		defer userLocks.Delete(id)

		var req UserJointActionRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req.ID = id

		response, err := endpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원 신체능력 /daf
// @Summary 회원 신체능력 조회
// @Description 회원 신체능력 조회시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} UserJointActionResponse "신체능력 정보"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-user-daf [get]
func GetUserHandler(endpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 토큰 검증 및 처리
		id, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response, err := endpoint(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(UserJointActionResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원별 운동 /daf
// @Summary 회원별 운동추천
// @Description 회원별 추천운동 조회시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} map[uint]RecomendResponse "카테고리아이디:추천운동"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-recommend [get]
func GetRecommendsHandler(endpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 토큰 검증 및 처리
		id, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response, err := endpoint(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(map[uint]RecomendResponse)
		c.JSON(http.StatusOK, resp)
	}
}
