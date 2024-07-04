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

func GetSuperEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getSuperAgencis()
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
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
		result, err := s.getAgencis(request.(uint))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetAdminsEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getAdmins(request.(uint))
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
		reqMap := request.(map[string]interface{})
		id := reqMap["id"].(uint)
		uid := reqMap["user_id"].(uint)
		result, err := s.getAfcs(id, uid)
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
		return BasicResponse{Code: result}, nil
	}
}

func UpdateAfcEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.updateAfc(request.(SaveAfcRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: result}, nil
	}
}

func GetAfcHistorisEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqMap := request.(map[string]interface{})
		id := reqMap["id"].(uint)
		uid := reqMap["user_id"].(uint)
		result, err := s.getAfcHistoris(id, uid)
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
		return BasicResponse{Code: result}, nil
	}
}

func SearhDiaryEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.searchDiary(request.(SearchDiaryRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func SaveDiaryEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.saveDiary(request.(SaveDiaryRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: result}, nil
	}
}

func GetExerciseMeasuresEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getExerciseMeasure()
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetAllUsersEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getAllUsers(request.(uint))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetUserEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqMap := request.(map[string]interface{})
		id := reqMap["id"].(uint)
		uid := reqMap["uid"].(uint)
		result, err := s.getUser(id, uid)
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func SearchMachinesEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.searchMachines(request.(SearchMachineRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func GetMachinesEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.getMachines(request.(uint))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return result, nil
	}
}

func SaveMachinesEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.saveMachines(request.(PostMachineRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: result}, nil
	}
}

func RemoveMachinesEndpoint(s AdminService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		result, err := s.removeMachines(request.(PostMachineRequest))
		if err != nil {
			return BasicResponse{Code: err.Error()}, err
		}
		return BasicResponse{Code: result}, nil
	}
}
