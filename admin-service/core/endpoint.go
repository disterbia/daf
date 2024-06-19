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

func ResetPasswordEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		info := request.(LoginRequest)
		code, err := s.resetPassword(info)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func SaveUserEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		saveUser := request.(SaveUserRequest)
		code, err := s.saveUser(saveUser)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: code}, nil
	}
}

func SearhUsersEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		param := request.(SearchUserRequest)
		result, err := s.searchUsers(param)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetAgencisEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getAgencis()
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetAdminsEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getAdmins()
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetDisableDetailsEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getDisableDetails()
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetAfcsEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getAfcs(request.(uint))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func CreateAfcEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.createAfc(request.(SaveAfcRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func UpdateAfcEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.updateAfc(request.(SaveAfcRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetAfcHistorisEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getAfcHistoris(request.(uint))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func UpdateAfcHistoryEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.updateAfcHistory(request.(SaveAfcHistoryRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}
