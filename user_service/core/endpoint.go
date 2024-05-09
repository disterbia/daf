package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func SnsLoginEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(LoginRequest)
		token, err := s.snsLogin(req)
		if err != nil {
			return LoginResponse{Err: err.Error()}, err
		}
		return LoginResponse{Jwt: token}, nil
	}
}

func AutoLoginEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		autoRequest := request.(AutoLoginRequest)
		token, err := s.autoLogin(autoRequest)
		if err != nil {
			return LoginResponse{Err: err.Error()}, err
		}
		return LoginResponse{Jwt: token}, nil
	}
}

func GetUserEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(uint)
		result, err := s.getUser(id)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func SetUserEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		user := request.(UserRequest)
		code, err := s.setUser(user)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func RemoveEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		uid := request.(uint)
		code, err := s.removeUser(uid)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func RemoveProfileEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(uint)
		code, err := s.removeProfile(id)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func GetVersionEndpoint(s UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		version, err := s.getVersion()
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return version, nil
	}
}
