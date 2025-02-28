package core

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-kit/kit/endpoint"
	"github.com/gofiber/fiber/v2"
)

var userLocks sync.Map

// @Tags 로그인 /user
// @Summary sns 로그인
// @Description sns 로그인 성공시 호출
// @Accept  json
// @Produce  json
// @Param request body LoginRequest true "요청 DTO"
// @Success 200 {object} SuccessResponse "성공시 JWT 토큰 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환: 오류메시지 "-1" = 인증필요 , "-2" = 이미 가입한 번호"
// @Router /sns-login [post]
func SnsLoginHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		response, err := endpoint(c.Context(), req)
		resp := response.(LoginResponse)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"jwt": resp.Jwt})
	}
}

// @Tags 로그인 /user
// @Summary 자동로그인
// @Description 최초 로그인 이후 앱 실행시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body AutoLoginRequest true "요청 DTO"
// @Success 200 {object} SuccessResponse "성공시 JWT 토큰 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Security jwt
// @Router /auto-login [post]
func AutoLoginHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 토큰 검증 및 처리
		_, email, err := verifyJWT(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		var req AutoLoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		req.Email = email
		response, err := endpoint(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		resp := response.(LoginResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 회원상태 변경(본인)  /user
// @Summary 유저 데이터 변경
// @Description 유저 상태영구변경시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Param request body UserRequest true "요청 DTO - 업데이트 할 데이터"
// @Success 200 {object} BasicResponse "성공시 200 반환"
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

		var req UserRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		req.ID = id

		response, err := endpoint(c.Context(), req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})

		}
		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 회원조회(본인)  /user
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

// @Tags 회원상태 변경(본인)  /user
// @Summary 프로필 사진 삭제
// @Description 기본이미지로 변경시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Security jwt
// @Router /remove-profile [post]
func RemoveProfileHandler(endpoint endpoint.Endpoint) fiber.Handler {
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

		resp := response.(BasicResponse)
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

// @Tags 회원탈퇴 /user
// @Summary 회원탈퇴
// @Description 회원탈퇴시 호출
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {jwt_token}"
// @Success 200 {object} BasicResponse "성공시 200 반환"
// @Failure 400 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Failure 500 {object} ErrorResponse "요청 처리 실패시 오류 메시지 반환"
// @Router /remvoe-user [post]
func RemoveHandler(endpoint endpoint.Endpoint) fiber.Handler {
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
		log.Println("jwt:", jwt)

		// ID 토큰을 웹 링크로 리다이렉트
		baseUrl := "https://localhost:64447/apple"
		redirectURL := fmt.Sprintf("%s?jwt=%s&code=%s", baseUrl, jwt, code)

		// 웹으로 리다이렉트 (302 리다이렉트)
		return c.Redirect(redirectURL, fiber.StatusFound)
	}
}

func GoogleCallbackHandler(endpoint endpoint.Endpoint) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ✅ GET 요청에서 code를 직접 가져옴 (BodyParser 제거)
		code := c.Query("code")
		state := c.Query("state")
		log.Println("code:", code, "state:", state)

		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "authorization code is missing",
			})
		}

		// ✅ `code`를 요청 구조체에 담아 전달
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
		jwtToken := resp.Jwt
		log.Println("JWT:", jwtToken)

		// ✅ 클라이언트로 리다이렉트
		baseUrl := "https://localhost:64447/google"
		redirectURL := fmt.Sprintf("%s?jwt=%s&code=%s", baseUrl, jwtToken, code)
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
		jwtToken := resp.Jwt
		log.Println("JWT:", jwtToken)

		// 클라이언트로 리다이렉트
		redirectURL := fmt.Sprintf("https://localhost:64447/kakao?jwt=%s&code=%s", jwtToken, code)
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
		jwtToken := resp.Jwt
		log.Println("JWT:", jwtToken)

		// 클라이언트로 리다이렉트
		redirectURL := fmt.Sprintf("https://localhost:64447/facebook?jwt=%s&code=%s", jwtToken, code)
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
		jwtToken := resp.Jwt
		log.Println("JWT:", jwtToken)

		// 클라이언트로 리다이렉트
		redirectURL := fmt.Sprintf("https://localhost:64447/naver?jwt=%s&code=%s", jwtToken, code)
		return c.Redirect(redirectURL, fiber.StatusFound)
	}
}
