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
// @Param Authorization header string true "Bearer {jwt_token}"
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

		// 사용자별 잠금 시작
		if _, loaded := userLocks.LoadOrStore(id, true); loaded {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Concurrent request detected"})
			return
		}
		defer userLocks.Delete(id)

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

// @Tags 회원검색 /admin
// @Summary 회원리스트 조회
// @Description 회원리스트 조회시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body SearchUserRequest true "요청 DTO page default=0, gender 1= 남 2=여 "
// @Success 200 {object} []SearchUserResponse "응답 DTO"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /search-users [post]
func SearchUsersHandler(searchEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var req SearchUserRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := searchEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.([]SearchUserResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원검색 /admin
// @Summary 등록매장 조회
// @Description 등록매장 조회시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} []AgAdResponse "응답 DTO"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-agencis [get]
func GetAgencisHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response, err := getEndpoint(c.Request.Context(), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.([]AgAdResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원검색 /admin
// @Summary 담당코치 조회
// @Description 담당코치 조회시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} []AgAdResponse "응답 DTO"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-admins [get]
func GetAdminsHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response, err := getEndpoint(c.Request.Context(), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.([]AgAdResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원검색 /admin
// @Summary 기타 장애유형 조회
// @Description 기타 장애유형시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} []AgAdResponse "응답 DTO"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-details [get]
func GetDisableDetailsHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response, err := getEndpoint(c.Request.Context(), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.([]AgAdResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원 Afc /admin
// @Summary 회원 Afc 조회
// @Description 현재 Afc조회시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param id path string true "id"
// @Success 200 {object} GetAfcResponse "응답 DTO"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-afcs/{id} [get]
func GetAfcsHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		userId := c.Param("id")
		response, err := getEndpoint(c.Request.Context(), userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(GetAfcResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원 Afc /admin
// @Summary 회원 Afc 등록
// @Description 회원 Afc 등록시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body SaveAfcRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /create-afc [post]
func CreateAfcHandler(myEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		var req SaveAfcRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.Id = id
		response, err := myEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원 Afc /admin
// @Summary 회원 Afc 수정
// @Description 현재 Afc 수정시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body SaveAfcRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /update-afc [post]
func UpdateAfcHandler(myEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		var req SaveAfcRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.Id = id
		response, err := myEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원 Afc /admin
// @Summary 회원 Afc history 조회
// @Description 회원 Afc history 조회시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param id path string true "id"
// @Success 200 {object} []GetAfcResponse "응답DTO"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-afc-historis/{id} [get]
func GetAfcHistorisHandler(myEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		userId := c.Param("id")
		response, err := myEndpoint(c.Request.Context(), userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.([]GetAfcResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 회원 Afc /admin
// @Summary 회원 Afc history 수정
// @Description 회원 Afc history 수정시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body SaveAfcHistoryRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /update-afc-history [post]
func UpdateAfcHistoryHandler(myEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		var req SaveAfcHistoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.Id = id

		response, err := myEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}
