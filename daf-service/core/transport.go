package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
	kitEndpoint "github.com/go-kit/kit/endpoint"
)

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
