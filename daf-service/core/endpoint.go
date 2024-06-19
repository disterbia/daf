package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func SetUserEndpoint(s DafService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		user := request.(UserAfcRequest)
		code, err := s.setUser(user)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func GetUserEndpoint(s DafService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(uint)
		res, err := s.getUser(id)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return res, nil
	}
}

func GetRecommendsEndpoint(s DafService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(uint)
		res, err := s.getRecommends(id)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return res, nil
	}
}
