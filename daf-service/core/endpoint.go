package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

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
