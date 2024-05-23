package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func LoginEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		info := request.(LoginRequest)
		token, err := s.login(info)
		if err != nil {
			return LoginResponse{Err: err.Error()}, err
		}
		return LoginResponse{Jwt: token}, nil
	}
}

func SendCodeEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		number := request.(string)
		code, err := s.sendAuthCode(number)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func VerifyEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		veri := request.(VerifyRequest)
		code, err := s.verifyAuthCode(veri)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func SignInEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		info := request.(SignInRequest)
		code, err := s.signIn(info)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}
