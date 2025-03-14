package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func GetSalesEndpoint(s PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetSalesRequest)
		res, err := s.getSales(req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func DeleteCartEndpoint(s PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteCartRequest)
		code, err := s.deleteCarts(req)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func CountCartEndpoint(s PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CountRequest)
		code, err := s.countCart(req)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func GetCartEndpoint(s PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(uint)
		res, err := s.getCart(id)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}
func SaveCartEndpoint(s PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqMap := request.(map[string]interface{})
		id := reqMap["uid"].(uint)
		productOptionId := reqMap["product_option_id"].(uint)
		code, err := s.saveCart(id, productOptionId)
		if err != nil {
			return nil, err
		}
		return BasicResponse{Code: code}, nil
	}
}

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
