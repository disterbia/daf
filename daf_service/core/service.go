// /user-service/service/service.go

package core

import (
	"daf_service/model"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type DafService interface {
	setUser(userRequest UserJointActionRequest) (string, error)
	getUser(id uint) (UserJointActionResponse, error)
}

type dafService struct {
	db *gorm.DB
}

func NewDafService(db *gorm.DB) DafService {
	return &dafService{db: db}
}

func (service *dafService) setUser(request UserJointActionRequest) (string, error) {

	if err := validateRequest(request); err != nil {
		return "", err
	}

	var jointActions []model.JointAction
	if err := service.db.Find(&jointActions).Error; err != nil {
		return "", errors.New("db error")
	}

	var clinicalFeatures []model.ClinicalFeature
	if err := service.db.Find(&clinicalFeatures).Error; err != nil {
		return "", errors.New("db error")
	}

	actions := []model.UserJointAction{}

	fields := reflect.TypeOf(request)
	values := reflect.ValueOf(request)

	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		value := values.Field(i)
		data := value.String()

		if field.Name == "ID" || field.Type.Kind() != reflect.String {
			continue
		}
		var clinicalId uint
		for _, jointAction := range jointActions {
			if jointAction.Name == field.Name {
				if data != "" {
					for _, clinicalFeature := range clinicalFeatures {
						if clinicalFeature.Code == strings.ToUpper(string(data[1])) {
							clinicalId = clinicalFeature.ID
						}
					}
					romId := uint(data[0] - '0')
					degreeId := uint(data[2] - '0')
					action := model.UserJointAction{
						Uid:               request.ID,
						JointActionId:     jointAction.ID,
						RomId:             romId,
						Name:              jointAction.Name,
						ClinicalFeatureId: clinicalId,
						DegreeId:          degreeId,
					}
					actions = append(actions, action)
				}
			}
		}

	}

	tx := service.db.Begin()
	if err := tx.Where("uid = ?", request.ID).Unscoped().Delete(&model.UserJointAction{}).Error; err != nil {
		return "", errors.New("db error")
	}

	if err := tx.Create(&actions).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error")
	}
	tx.Commit()
	return "200", nil
}

func (service *dafService) getUser(id uint) (UserJointActionResponse, error) {

	var userJointActions []model.UserJointAction
	if err := service.db.Where("uid = ?", id).Preload("JointAction").Preload("ClinicalFeature").Find(&userJointActions).Error; err != nil {
		return UserJointActionResponse{}, errors.New("db error")
	}

	var userJointActionResponse UserJointActionResponse
	fields := reflect.TypeOf(userJointActionResponse)
	responseValue := reflect.ValueOf(&userJointActionResponse).Elem()

	type GroupData struct {
		romList    uint
		clinicList []string
		degreeList uint
		count      uint
	}
	groupData := make(map[uint]GroupData)

	for _, userJointAction := range userJointActions {
		for i := 0; i < fields.NumField(); i++ {
			field := fields.Field(i)
			if userJointAction.JointAction.Name == field.Name {

				romId := userJointAction.RomId
				clinicalFeture := userJointAction.ClinicalFeature.Code
				degreeId := userJointAction.DegreeId
				resultCode := strconv.FormatUint(uint64(romId), 10) + clinicalFeture + strconv.FormatUint(uint64(degreeId), 10)

				// 필드가 유효한지 확인 후 설정
				if responseField := responseValue.FieldByName(field.Name); responseField.IsValid() && responseField.CanSet() {
					responseField.Set(reflect.ValueOf(resultCode))
				} else {
					fmt.Printf("Invalid field: %s\n", field.Name) // 디버깅 정보 추가
				}

			}
		}
		// BodyCompositionId를 키로 사용하는 그룹에 데이터 추가
		bodyCompId := userJointAction.JointAction.BodyCompositionId
		data := groupData[bodyCompId]
		data.romList += userJointAction.RomId
		data.clinicList = append(data.clinicList, userJointAction.ClinicalFeature.Code)
		data.degreeList += userJointAction.DegreeId
		data.count++
		groupData[bodyCompId] = data
	}

	// 그룹별 평균 계산
	for bodyCompId, data := range groupData {
		romAver := data.romList / data.count
		degreeAver := data.degreeList / data.count
		var clinicAver string

		// 빈도 수를 기록하기 위한 해시맵
		frequency := make(map[string]int)
		// 가장 많은 문자열과 그 빈도 수를 추적
		maxCount := 0

		// 각 문자열의 빈도 수를 해시맵에 기록하고 가장 많은 문자열을 찾기
		for _, str := range data.clinicList {
			frequency[str]++
			if frequency[str] > maxCount {
				clinicAver = str
				maxCount = frequency[str]
			}
		}
		resultCode := strconv.FormatUint(uint64(romAver), 10) + clinicAver + strconv.FormatUint(uint64(degreeAver), 10)
		switch bodyCompId {
		// responseValue
		case uint(UL):
			responseValue.FieldByName("ULAV").Set(reflect.ValueOf(resultCode))
		case uint(UR):
			responseValue.FieldByName("URAV").Set(reflect.ValueOf(resultCode))
		case uint(LL):
			responseValue.FieldByName("LLAV").Set(reflect.ValueOf(resultCode))
		case uint(LR):
			responseValue.FieldByName("LRAV").Set(reflect.ValueOf(resultCode))
		case uint(TR):
			responseValue.FieldByName("TR").Set(reflect.ValueOf(resultCode))
		}

		fmt.Printf("BodyCompositionId: %d, ROM 평균: %d, Degree 평균: %d, Clinic 평균: %s\n", bodyCompId, romAver, degreeAver, clinicAver)
	}
	return responseValue.Interface().(UserJointActionResponse), nil
}
