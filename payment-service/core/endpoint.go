package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func PaymentCallbackEndpoint(s PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PaymentCallbackResponse)
		code, err := s.paymentCallback(req)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func RefundEndpoint(s PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		code, err := s.refund()
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}
