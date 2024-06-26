package core

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	kitEndpoint "github.com/go-kit/kit/endpoint"

	"github.com/gin-gonic/gin"
)

var userLocks sync.Map

// @Tags 코치로그인 /coach
// @Summary 로그인
// @Description 로그인시 호출
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

// @Tags 코치 /coach
// @Summary 카테고리 생성/수정
// @Description 최상위 운동 카테고리 생성/수정시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body CategoryRequest true "요청 DTO - 생성시 id 생략"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /save-category [post]
func SaveCategoryHandler(saveEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
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

		var req CategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := saveEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 카테고리 조회
// @Description 카테고리 조회시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} []CategoryResponse "카테고리 정보"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-categoris [get]
func GetCategorisHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
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

		resp := response.([]CategoryResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 운동 생성/수정
// @Description 운동 생성/수정시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body ExerciseRequest true "요청 DTO - 생성시 id 생략"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /save-exercise [post]
func SaveExerciseHandler(saveEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
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

		var req ExerciseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := saveEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 기구 조회
// @Description 기구 조회시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} []MachineDto "기구 정보"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-machines [get]
func GetMachinesHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
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

		resp := response.([]MachineDto)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 기구 생성/수정
// @Description 기구 생성/수정시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body MachineDto true "요청 DTO - 생성시 id 생략"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /save-machine [post]
func SaveMachineHandler(saveEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
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

		var req MachineDto
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := saveEndpoint(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 운동목적 조회
// @Description 운동목적 리스트 조회시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} []PurposeDto "운동목적 리스트"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-purposes [get]
func GetPurposesHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
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

		resp := response.([]PurposeDto)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 추천 운동 생성/수정
// @Description 추천 운동 생성/수정시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body RecommendRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /save-recommend [post]
func SaveRecommendHandler(saveEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}

		// 사용자별 잠금 시작
		if _, loaded := userLocks.LoadOrStore(id, true); loaded {
			c.JSON(http.StatusTooManyRequests, gin.H{"err": "Concurrent request detected"})
			return
		}
		defer userLocks.Delete(id)

		var req RecommendRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}

		response, err := saveEndpoint(c.Request.Context(), req)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}

		resp := response.(BasicResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 운동 상세조회
// @Description 운동 상세 조회시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param id path string true "exercise_id"
// @Success 200 {object} RecommendResponse "운동 정보"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-exercise/{exercise_id} [get]
func GetRecommendHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		exerciseId := c.Param("exercise_id")
		id, err := strconv.Atoi(exerciseId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := getEndpoint(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.(RecommendResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 추천 운동 전체 조회
// @Description 최초 운동리스트 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} []RecommendResponse "운동 리스트"
// @Param  page  query uint false  "페이지 번호 default 0 30개씩"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-recommends [get]
func GetRecommendsHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		pageParam := c.Query("page")
		var page uint
		if pageParam != "" {
			parsed, err := strconv.ParseUint(pageParam, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
				return
			}
			page = uint(parsed)
		}

		response, err := getEndpoint(c.Request.Context(), page)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.([]RecommendResponse)
		c.JSON(http.StatusOK, resp)
	}
}

// @Tags 코치 /coach
// @Summary 추천 운동 검색
// @Description 검색시 호출
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} []RecommendResponse "운동 리스트"
// @Param  page  query uint false  "페이지 번호 default 0 30개씩"
// @Param  name  query string false "검색명"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /search-recommends [get]
func SearchRecommendsHandler(getEndpoint kitEndpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, err := verifyJWT(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var queryParams SearchRequest
		if err := c.ShouldBindQuery(&queryParams); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := getEndpoint(c.Request.Context(), queryParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := response.([]RecommendResponse)
		c.JSON(http.StatusOK, resp)
	}
}
