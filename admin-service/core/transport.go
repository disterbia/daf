package core

import (
	"net/http"
	"sync"

	kitEndpoint "github.com/go-kit/kit/endpoint"

	"github.com/gin-gonic/gin"
)

var userLocks sync.Map

// @Tags 관리자 로그인 /admin
// @Summary 관리자 로그인
// @Description 관리자 로그인시 호출
// @Accept  json
// @Produce  json
// @Param email body string true "email"
// @Param password body string true "password"
// @Success 200 {object} SuccessResponse "성공시 JWT 토큰 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Security jwt
// @Router /login [post]
func LoginHandler(loginEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := loginEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(LoginResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 인증번호 /admin
// @Summary 인증번호 발송
// @Description 인증번호 발송시 호출
// @Accept  json
// @Produce  json
// @Param email path string true "이메일"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /send-code/{email} [post]
func SendCodeHandler(sendEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {

		number := c.Param("email")

		response, err := sendEndpoint(c.Request.Context(), number)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 인증번호 /admin
// @Summary 번호 인증
// @Description 인증번호 입력 후 호출
// @Accept  json
// @Produce  json
// @Param request body VerifyRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /verify-code [post]
func VerifyHandler(verifyEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req VerifyRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := verifyEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 관리자 회원가입 /admin
// @Summary 회원가입
// @Description 회원 가입시 호출
// @Accept  json
// @Produce  json
// @Param request body SignInRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /sign-in [post]
func SignInHandler(siginEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req SignInRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := siginEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 관리자 비밀번호 재설정 /admin
// @Summary 비밀번호 재설정
// @Description 비밀번호 재설정시 호출
// @Accept  json
// @Produce  json
// @Param request body LoginRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /reset-password [post]
func ResetPasswordHandler(resetEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req LoginRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := resetEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원등록 /admin
// @Summary 회원등록
// @Description 회원등록시 호출
// @Accept  json
// @Produce  json
// @Param request body SaveUserRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /save-user [post]
func SaveUserHandler(saveEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var req SaveUserRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req.Uid = id
		response, err := saveEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}
