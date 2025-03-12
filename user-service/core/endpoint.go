package core

import (
	"context"
	"errors"
	"log"

	"github.com/go-kit/kit/endpoint"
)

func BaiscLoginEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(LoginRequest)
		token, err := s.basicLogin(req)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: token}, nil
	}
}

func CheckUsernameEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(string)
		code, err := s.checkUsername(req)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func GetUserEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(uint)
		result, err := s.getUser(id)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}

func VerifyEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		veri := request.(VerifyRequest)
		code, err := s.verifyAuthCode(veri.PhoneNumber, veri.Code)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func PaymentCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PaymentCallbackResponse)
		code, err := s.paymentCallback(req)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func RefundEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		code, err := s.refund()
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func SendCodeEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		number := request.(string)
		code, err := s.sendAuthCode(number)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func SignInEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		dto := request.(SignInRequest)
		code, err := s.signIn(dto)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func SetUserEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		dto := request.(SetUserRequest)
		code, err := s.setUser(dto)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func FinUsernameEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		dto := request.(FindUsernameRequest)
		code, err := s.findUsername(dto)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func FindPasswordEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		dto := request.(FindPasswordRequest)
		code, err := s.findPassword(dto)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func AppleCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			log.Println("code")
			return nil, errors.New("authorization code is missing")
		}

		// Authorization Code로 애플과 통신해 토큰을 교환합니다.
		response, err := s.appleLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return response, nil

	}
}

func GoogleCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			log.Println("code is missing")
			return nil, errors.New("authorization code is missing")
		}

		response, err := s.googleLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
}
func KakaoCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			log.Println("code is missing")
			return nil, errors.New("authorization code is missing")
		}

		response, err := s.kakaoLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
}

func FacebookCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			return nil, errors.New("authorization code is missing")
		}

		// Authorization Code로 Access Token 요청
		response, err := s.facebookLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
}
func NaverCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			return nil, errors.New("authorization code is missing")
		}

		response, err := s.naverLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
}
