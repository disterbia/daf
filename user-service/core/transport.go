package core

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

var userLocks sync.Map
var ipLimiters = make(map[string]*rate.Limiter)
var ipLimitersMutex sync.Mutex

func getClientIP(c *fiber.Ctx) string {
	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return c.IP()
}

// @Tags 회원수정 /user
// @Summary 회원 데이터 변경
// @Description 회원 정보 변경 시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body SetUserRequest true "요청 DTO - 업데이트 할 데이터/password 필드가 빈값이 아닐때 비밀번호만 업데이트"
// @Success 200 {object} BasicResponse "성공시 1 / -1: 번호 인증안됨"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /set-user [post]
func SetUserHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 토큰 검증 및 처리
		id, _, err := verifyJWT(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		var req SetUserRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		req.Uid = id

		response, err := endpoint(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})

		}
		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 아이디 찾기 /user
// @Summary 아이디 찾기
// @Description 아이디찾기 시 호출
// @Accept  json
// @Produce  json
// @Param request body FindUsernameRequest true "정보 dto"
// @Success 200 {object} BasicResponse "성공시 아이디 반환/ -1: 번호 인증안됨 / -2: 일치하는 회원 없음"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /find-username [post]
func FindUsernameHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req FindUsernameRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		response, err := endpoint(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})

		}
		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 비밀번호 찾기 /user
// @Summary 비밀번호 찾기
// @Description 비밀번호 찾기 시 호출
// @Accept  json
// @Produce  json
// @Param request body FindPasswordRequest true "정보 dto"
// @Success 200 {object} BasicResponse "성공시 JWT 토큰 반환/ -1: 번호 인증안됨 / -2: 일치하는 회원 없음"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /find-password [post]
func FindPasswordHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req FindPasswordRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		response, err := endpoint(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})

		}

		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 회원조회 /user
// @Summary 유저 조회
// @Description 내 정보 조회시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} UserResponse "성공시 유저 객체 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /get-user [post]
func GetUserHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 토큰 검증 및 처리
		id, _, err := verifyJWT(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		response, err := endpoint(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		resp := response.(UserResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 회원가입 /user
// @Summary 중복확인
// @Description 아이디 중복확인 시 호출
// @Accept  json
// @Produce  json
// @Param username query string true "중복체크 할 아이디"
// @Success 200 {object} BasicResponse "성공시 1,이미 있는 아이디 -1"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /check-username [get]
func CheckUsernameHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		username := c.Query("username")
		response, err := endpoint(context.Background(), username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		resp := response.(BasicResponse)

		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 로그인 /user
// @Summary 일반로그인
// @Description 아이디/비밀번호 로그인 시 호출
// @Accept  json
// @Produce  json
// @Param request body LoginRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 JWT 토큰 반환/-1 :아이디 또는 비밀번호 불일치"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /login [post]
func BasicLoginHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		response, err := endpoint(context.Background(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		resp := response.(BasicResponse)

		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 회원가입 /user
// @Summary 회원가입
// @Description 회원가입 정보 입력 완료 후 호출
// @Accept  json
// @Produce  json
// @Param request body SignInRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 1, 휴대폰 인증 안함 -1, 추천인 없음 -2"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환 "
// @Router /sign-in [post]
func SignInHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SignInRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		response, err := endpoint(context.Background(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		resp := response.(BasicResponse)

		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 인증번호 /user
// @Summary 인증번호 발송
// @Description 휴대전화 인증번호 발송시 호출
// @Accept  json
// @Produce  json
// @Param number path string true "휴대번호"
// @Success 200 {object} BasicResponse "성공시 1 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /send-code/{number} [post]
func SendCodeHandler(sendEndpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		number := c.Params("number")

		// IP 주소를 가져오기 위한 함수 호출
		ip := getClientIP(c)

		ipLimitersMutex.Lock()
		limiter, exists := ipLimiters[ip]
		if !exists {
			limiter = rate.NewLimiter(rate.Every(24*time.Hour), 10)
			ipLimiters[ip] = limiter
		}
		ipLimitersMutex.Unlock()

		// 요청이 허용되지 않으면 에러 반환
		if !limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "요청 횟수 초과"})
		}
		response, err := sendEndpoint(c.Context(), number)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// 응답이 성공적이면 RateLimiter를 업데이트
		ipLimitersMutex.Lock()
		limiter.Allow()
		ipLimitersMutex.Unlock()

		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 인증번호 /user
// @Summary 인증번호 인증
// @Description 인증번호 인증시 호출
// @Accept  json
// @Produce  json
// @Param request body VerifyRequest true "요청 DTO"
// @Success 200 {object} BasicResponse "성공시 1 반환 코드불일치 -1"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /verify-code [post]
func VerifyHandler(verifyEndpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var req VerifyRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		response, err := verifyEndpoint(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

func AppleCallbackHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Query("code")
		state := c.Query("state")
		log.Println("code:", code, "state:", state)
		// POST 요청에서 body 파싱
		var req CallbackRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		if req.Code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "authorization code is missing",
			})
		}

		// 엔드포인트 호출
		response, err := endpoint(context.Background(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// 응답에서 ID 토큰 추출
		resp := response.(LoginResponse)
		jwt := resp.Jwt
		snsId := resp.SnsId

		// ID 토큰을 웹 링크로 리다이렉트
		baseUrl := "http://192.168.0.24:59704/apple"
		redirectURL := fmt.Sprintf("%s?jwt=%s&code=%s&sns_id=%s&sns_email=%s", baseUrl, jwt, code, snsId, resp.SnsEmail)

		// 웹으로 리다이렉트 (302 리다이렉트)
		return c.Redirect(redirectURL, fiber.StatusFound)
	}
}

func GoogleCallbackHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// GET 요청에서 code를 직접 가져옴 (BodyParser 제거)
		code := c.Query("code")
		state := c.Query("state")
		log.Println("code:", code, "state:", state)

		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "authorization code is missing",
			})
		}

		// `code`를 요청 구조체에 담아 전달
		req := CallbackRequest{Code: code, State: state}

		// 엔드포인트 호출
		response, err := endpoint(context.Background(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// 응답에서 JWT 추출
		resp := response.(LoginResponse)
		jwt := resp.Jwt
		snsId := resp.SnsId

		// 클라이언트로 리다이렉트
		baseUrl := "http://192.168.0.24:59704/google"
		redirectURL := fmt.Sprintf("%s?jwt=%s&code=%s&sns_id=%s&sns_email=%s", baseUrl, jwt, code, snsId, resp.SnsEmail)
		return c.Redirect(redirectURL, fiber.StatusFound)
	}
}

func KakaoCallbackHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Query("code") // 카카오에서 받은 Authorization Code
		log.Println("Kakao Code:", code)

		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "authorization code is missing",
			})
		}

		// 엔드포인트 호출
		response, err := endpoint(context.Background(), CallbackRequest{Code: code})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// 응답에서 JWT 추출
		resp := response.(LoginResponse)
		jwt := resp.Jwt
		snsId := resp.SnsId

		// 클라이언트로 리다이렉트
		redirectURL := fmt.Sprintf("http://192.168.0.24:59704/kakao?jwt=%s&code=%s&sns_id=%s&sns_email=%s", jwt, code, snsId, resp.SnsEmail)
		return c.Redirect(redirectURL, fiber.StatusFound)
	}
}

func FacebookCallbackHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Query("code") // 페이스북에서 받은 Authorization Code
		log.Println("Facebook Code:", code)

		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "authorization code is missing",
			})
		}

		// 엔드포인트 호출
		response, err := endpoint(context.Background(), CallbackRequest{Code: code})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// 응답에서 JWT 추출
		resp := response.(LoginResponse)
		jwt := resp.Jwt
		snsId := resp.SnsId

		// 클라이언트로 리다이렉트
		redirectURL := fmt.Sprintf("http://192.168.0.24:59704/facebook?jwt=%s&code=%s&sns_id=%s&sns_email=%s", jwt, code, snsId, resp.SnsEmail)
		return c.Redirect(redirectURL, fiber.StatusFound)
	}
}

func NaverCallbackHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Query("code")
		state := c.Query("state")
		log.Println("Naver Code:", code, "State:", state)

		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "authorization code is missing",
			})
		}

		response, err := endpoint(context.Background(), CallbackRequest{Code: code, State: state})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// 응답에서 JWT 추출
		resp := response.(LoginResponse)
		jwt := resp.Jwt
		snsId := resp.SnsId

		// 클라이언트로 리다이렉트
		redirectURL := fmt.Sprintf("http://192.168.0.24:59704/naver?jwt=%s&code=%s&sns_id=%s&sns_email=%s", jwt, code, snsId, resp.SnsEmail)
		return c.Redirect(redirectURL, fiber.StatusFound)
	}
}
