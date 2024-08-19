package core

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func LoginEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		info := request.(LoginRequest)
		token, err := s.login(info)
		if err != nil {
			return LoginResponse{Err: err.Error()}, err
		}
		return LoginResponse{Jwt: token}, nil
	}
}

func GetCategorisEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		categoris, err := s.getCategoris()
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return categoris, nil
	}
}

func SaveCategoryEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		dto := request.(CategoryRequest)
		msg, err := s.saveCategory(dto.ID, dto.Name)
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return BasicResponse{Msg: msg}, nil
	}
}

func SaveExerciseEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		msg, err := s.saveExercise(request.(ExerciseRequest))
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return BasicResponse{Msg: msg}, nil
	}
}

func GetMachinesEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		machines, err := s.getMachines()
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return machines, nil
	}
}

func SaveMachineEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		msg, err := s.saveMachine(request.(MachineDto))
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return BasicResponse{Msg: msg}, nil
	}
}

func SavePurposeEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		msg, err := s.savePurpose(request.(PurposeDto))
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return BasicResponse{Msg: msg}, nil
	}
}

func GetPurposesEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		purposes, err := s.getPurposes()
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return purposes, nil
	}
}

func SaveRecommendEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		dto := request.(RecommendRequest)
		msg, err := s.saveRecommend(dto)
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return BasicResponse{Msg: msg}, nil
	}
}

func GetRecommendEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		exerciseId := request.(uint)
		recommend, err := s.getRecommend(exerciseId)
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return recommend, nil
	}
}

// func GetRecommendsEndpoint(s CoachService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		recommends, err := s.getRecommends(request.(uint))
// 		if err != nil {
// 			return BasicResponse{Msg: err.Error()}, err
// 		}
// 		return recommends, nil
// 	}
// }

func SearchRecommendsEndpoint(s CoachService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		dto := request.(SearchRequest)
		recommends, err := s.searchRecommend(dto.Page, dto.Name)
		if err != nil {
			return BasicResponse{Msg: err.Error()}, err
		}
		return recommends, nil
	}
}
