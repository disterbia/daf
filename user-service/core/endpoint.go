package core

import (
	"context"
	"errors"
	"log"

	"github.com/go-kit/kit/endpoint"
)

// func SnsLoginEndpoint(s UserService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		req := request.(LoginRequest)
// 		token, err := s.snsLogin(req)
// 		if err != nil {
// 			return LoginResponse{Err: err.Error()}, err
// 		}
// 		return LoginResponse{Jwt: token}, nil
// 	}
// }

// func GetUserEndpoint(s UserService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		id := request.(uint)
// 		result, err := s.getUser(id)
// 		if err != nil {
// 			return BasicResponse{Code: err.Error()}, err
// 		}
// 		return result, nil
// 	}
// }

// func SetUserEndpoint(s UserService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		user := request.(UserRequest)
// 		code, err := s.setUser(user)
// 		if err != nil {
// 			return BasicResponse{Code: err.Error()}, err
// 		}
// 		return BasicResponse{Code: code}, nil
// 	}
// }

// func RemoveEndpoint(s UserService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		uid := request.(uint)
// 		code, err := s.removeUser(uid)
// 		if err != nil {
// 			return BasicResponse{Code: err.Error()}, err
// 		}
// 		return BasicResponse{Code: code}, nil
// 	}
// }

//	func RemoveProfileEndpoint(s UserService) endpoint.Endpoint {
//		return func(ctx context.Context, request interface{}) (interface{}, error) {
//			id := request.(uint)
//			code, err := s.removeProfile(id)
//			if err != nil {
//				return BasicResponse{Code: err.Error()}, err
//			}
//			return BasicResponse{Code: code}, nil
//		}
//	}
func AppleCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			log.Println("code")
			return nil, errors.New("authorization code is missing")
		}

		// Authorization Code로 애플과 통신해 토큰을 교환합니다.
		token, err := s.appleLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return LoginResponse{Jwt: token}, nil

	}
}

func GoogleCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			log.Println("code is missing")
			return nil, errors.New("authorization code is missing")
		}

		token, err := s.googleLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return LoginResponse{Jwt: token}, nil
	}
}
func KakaoCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			log.Println("code is missing")
			return nil, errors.New("authorization code is missing")
		}

		token, err := s.kakaoLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return LoginResponse{Jwt: token}, nil
	}
}

func FacebookCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			return nil, errors.New("authorization code is missing")
		}

		// Authorization Code로 Access Token 요청
		token, err := s.facebookLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return LoginResponse{Jwt: token}, nil
	}
}
func NaverCallbackEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CallbackRequest)

		if req.Code == "" {
			return nil, errors.New("authorization code is missing")
		}

		token, err := s.naverLogin(req.Code)
		if err != nil {
			return nil, err
		}
		return LoginResponse{Jwt: token}, nil
	}
}
