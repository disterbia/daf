// /user-service/service/service.go

package core

import (
	"daf-service/model"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type DafService interface {
	setUser(userRequest UserAfcRequest) (string, error)
	getUser(id uint) (UserAfcResponse, error)
	getRecommends(id uint) (map[uint]RecomendResponse, error)
}

type dafService struct {
	db *gorm.DB
}

func NewDafService(db *gorm.DB) DafService {
	return &dafService{db: db}
}

func (service *dafService) setUser(request UserAfcRequest) (string, error) {

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

	actions := []model.UserAfc{}

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
					if field.Name != "TR" && field.Name != "LOCOMOTION" {
						for _, clinicalFeature := range clinicalFeatures {
							if clinicalFeature.Code == strings.ToUpper(string(data[1])) {
								clinicalId = clinicalFeature.ID
							}
						}
						romId := uint(data[0] - '0')
						degreeId := uint(data[2] - '0')
						action := model.UserAfc{
							Uid:               request.ID,
							JointActionID:     jointAction.ID,
							RomID:             romId,
							Name:              jointAction.Name,
							ClinicalFeatureID: &clinicalId,
							DegreeID:          &degreeId,
						}
						actions = append(actions, action)
					} else {
						action := model.UserAfc{
							Uid:               request.ID,
							JointActionID:     jointAction.ID,
							RomID:             uint(data[0] - '0'),
							Name:              jointAction.Name,
							ClinicalFeatureID: nil,
							DegreeID:          nil,
						}
						actions = append(actions, action)
					}

				}
			}
		}

	}

	tx := service.db.Begin()
	if err := tx.Where("uid = ?", request.ID).Unscoped().Delete(&model.UserAfc{}).Error; err != nil {
		return "", errors.New("db error")
	}

	if err := tx.Create(&actions).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error")
	}
	tx.Commit()
	return "200", nil
}

func (service *dafService) getUser(id uint) (UserAfcResponse, error) {

	var userJointActions []model.UserAfc
	if err := service.db.Where("uid = ?", id).Preload("JointAction").Preload("ClinicalFeature").Find(&userJointActions).Error; err != nil {
		return UserAfcResponse{}, errors.New("db error")
	}

	var userJointActionResponse UserAfcResponse
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
				if field.Name != "TR" && field.Name != "LOCOMOTION" {
					romId := userJointAction.RomID
					clinicalFeture := userJointAction.ClinicalFeature.Code
					degreeId := userJointAction.DegreeID
					resultCode := strconv.FormatUint(uint64(romId), 10) + clinicalFeture + strconv.FormatUint(uint64(*degreeId), 10)

					// 필드가 유효한지 확인 후 설정
					if responseField := responseValue.FieldByName(field.Name); responseField.IsValid() && responseField.CanSet() {
						responseField.Set(reflect.ValueOf(resultCode))
					} else {
						fmt.Printf("Invalid field: %s\n", field.Name) // 디버깅 정보 추가
					}
				} else {
					if responseField := responseValue.FieldByName(field.Name); responseField.IsValid() && responseField.CanSet() {
						romIdString := strconv.Itoa(int(userJointAction.RomID))
						responseField.Set(reflect.ValueOf(romIdString))
					} else {
						log.Println("Invalid field:", field.Name)

					}
				}

			}
		}
		// BodyCompositionID를 키로 사용하는 그룹에 데이터 추가
		bodyCompId := userJointAction.BodyCompositionID
		data := groupData[bodyCompId]
		data.romList += userJointAction.RomID
		if userJointAction.ClinicalFeatureID != nil && userJointAction.DegreeID != nil {
			data.clinicList = append(data.clinicList, userJointAction.ClinicalFeature.Code)
			data.degreeList += *userJointAction.DegreeID
		}
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
			responseValue.FieldByName("TR").Set(reflect.ValueOf(strconv.Itoa(int(romAver))))
		case uint(LOCOMOTION):
			responseValue.FieldByName("LOCOMOTION").Set(reflect.ValueOf(strconv.Itoa(int(romAver))))
		}

		fmt.Printf("BodyCompositionID: %d, ROM 평균: %d, Degree 평균: %d, Clinic 평균: %s\n", bodyCompId, romAver, degreeAver, clinicAver)
	}
	return responseValue.Interface().(UserAfcResponse), nil
}

func (service *dafService) getRecommends(id uint) (map[uint]RecomendResponse, error) {

	//유저 부위별 코드 가져오기
	var userJointActions []model.UserAfc
	if err := service.db.Where("uid = ?", id).Preload("JointAction").Preload("ClinicalFeature").Find(&userJointActions).Error; err != nil {
		return nil, errors.New("db error")
	}

	type GroupData struct {
		romList    uint
		clinicList []uint
		degreeList uint
		count      uint
	}
	groupData := make(map[uint]GroupData)

	type SearchData struct {
		bodyType uint
		rom      uint
		clinic   uint
		degree   uint
	}
	var ulav, urav, llav, lrav, tr, loco SearchData
	for _, userJointAction := range userJointActions {

		// BodyCompositionID를 키로 사용하는 그룹에 데이터 추가
		bodyCompId := userJointAction.BodyCompositionID
		data := groupData[bodyCompId]
		data.romList += userJointAction.RomID
		if userJointAction.ClinicalFeatureID != nil && userJointAction.DegreeID != nil {
			data.clinicList = append(data.clinicList, *userJointAction.ClinicalFeatureID)
			data.degreeList += *userJointAction.DegreeID
		}
		data.count++
		groupData[bodyCompId] = data
	}

	// 그룹별 평균 계산
	for bodyCompId, data := range groupData {
		romAver := data.romList / data.count
		degreeAver := data.degreeList / data.count
		var clinicAver uint

		// 빈도 수를 기록하기 위한 해시맵
		frequency := make(map[uint]int)
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
		switch bodyCompId {
		// responseValue
		case uint(UL):
			ulav = SearchData{bodyType: uint(UBODY), rom: romAver, clinic: clinicAver, degree: degreeAver}
		case uint(UR):
			urav = SearchData{bodyType: uint(UBODY), rom: romAver, clinic: clinicAver, degree: degreeAver}
		case uint(LL):
			llav = SearchData{bodyType: uint(LBODY), rom: romAver, clinic: clinicAver, degree: degreeAver}
		case uint(LR):
			lrav = SearchData{bodyType: uint(LBODY), rom: romAver, clinic: clinicAver, degree: degreeAver}
		case uint(TR):
			tr = SearchData{bodyType: uint(TBODY), rom: romAver}
		case uint(LOCOMOTION):
			loco = SearchData{bodyType: uint(LOCOBODY), rom: romAver}
		}

		fmt.Printf("BodyCompositionID: %d, ROM 평균: %d, Degree 평균: %d, Clinic 평균: %d\n", bodyCompId, romAver, degreeAver, clinicAver)
	}

	// var recommends []model.Recommended
	var recommends []model.Recommended
	err := service.db.Where(`
	(body_type_id = ? ) OR
	(body_type_id = ? ) OR
	(body_type_id = ? AND clinical_feature_id = ? AND degree_id <= ?) OR
	(body_type_id = ? AND clinical_feature_id = ? AND degree_id <= ?) OR
	(body_type_id = ? AND clinical_feature_id = ? AND degree_id <= ?) OR
	(body_type_id = ? AND clinical_feature_id = ? AND degree_id <= ?)`,
		tr.bodyType,
		loco.bodyType,
		ulav.bodyType, ulav.clinic, ulav.degree,
		urav.bodyType, urav.clinic, urav.degree,
		llav.bodyType, llav.clinic, llav.degree,
		lrav.bodyType, lrav.clinic, lrav.degree).
		Find(&recommends).Error
	if err != nil {
		return nil, errors.New("db error")
	}

	var recommendsTr, recommendsLoco, recommendsUl, recommendsUr, recommendsLl, recommendsLr []model.Recommended

	var originMap = make(map[uint][]model.Recommended)
	var recommendMap = make(map[uint][]model.Recommended)
	var searchMap = make(map[uint]SearchData)
	for _, rec := range recommends {
		switch {
		case rec.BodyTypeID == tr.bodyType:
			originMap[uint(TR)] = append(originMap[uint(TR)], rec)
			if rec.RomID <= tr.rom {
				recommendsTr = append(recommendsTr, rec)
				recommendMap[uint(TR)] = recommendsTr
			}

		case rec.BodyTypeID == loco.bodyType:
			originMap[uint(LOCOMOTION)] = append(originMap[uint(LOCOMOTION)], rec)
			if rec.RomID <= loco.rom {
				recommendsLoco = append(recommendsLoco, rec)
				recommendMap[uint(LOCOMOTION)] = recommendsLoco
			}

		case rec.BodyTypeID == ulav.bodyType:
			originMap[uint(UL)] = append(originMap[uint(UL)], rec)
			if rec.ClinicalFeatureID == ulav.clinic && rec.RomID <= ulav.rom && rec.DegreeID <= ulav.degree {
				recommendsUl = append(recommendsUl, rec)
				recommendMap[uint(UL)] = recommendsUl
				searchMap[uint(UL)] = SearchData{bodyType: ulav.bodyType, rom: ulav.rom, clinic: ulav.clinic, degree: ulav.degree}
			}

		case rec.BodyTypeID == urav.bodyType:
			originMap[uint(UR)] = append(originMap[uint(UR)], rec)
			if rec.ClinicalFeatureID == urav.clinic && rec.RomID <= urav.rom && rec.DegreeID <= urav.degree {
				recommendsUr = append(recommendsUr, rec)
				recommendMap[uint(UR)] = recommendsUr
				searchMap[uint(UR)] = SearchData{bodyType: urav.bodyType, rom: urav.rom, clinic: urav.clinic, degree: urav.degree}
			}

		case rec.BodyTypeID == llav.bodyType:
			originMap[uint(LL)] = append(originMap[uint(LL)], rec)
			if rec.ClinicalFeatureID == llav.clinic && rec.RomID <= llav.rom && rec.DegreeID <= llav.degree {
				recommendsLl = append(recommendsLl, rec)
				recommendMap[uint(LL)] = recommendsLl
				searchMap[uint(LL)] = SearchData{bodyType: llav.bodyType, rom: llav.rom, clinic: llav.clinic, degree: llav.degree}
			}

		case rec.BodyTypeID == lrav.bodyType:
			originMap[uint(LR)] = append(originMap[uint(LR)], rec)
			if rec.ClinicalFeatureID == lrav.clinic && rec.RomID <= lrav.rom && rec.DegreeID <= lrav.degree {
				recommendsLr = append(recommendsLr, rec)
				recommendMap[uint(LR)] = recommendsLr
				searchMap[uint(LR)] = SearchData{bodyType: lrav.bodyType, rom: lrav.rom, clinic: lrav.clinic, degree: lrav.degree}
			}
		}
	}

	for key, value := range recommendMap {
		if len(value) == 0 {
			if len(originMap[key]) == 0 {
				log.Println("최종적으로 할수있는 운동 없음.")
				return nil, nil
			}
			log.Println("적합한 운동이 없어서 ", key, "번 bodycomposition rom 조건 포함안함.")
			recommendMap[key] = append(recommendMap[key], originMap[key]...)
		}
	}

	exerciseIDMap := make(map[uint]int)

	for _, value := range recommendMap {
		for _, v := range value {
			exerciseIDMap[v.ExerciseID]++
		}
	}

	var commonExerciseIDs []uint
	for id, count := range exerciseIDMap {
		if count == 6 { // 모든 슬라이스에 포함된 ExerciseID
			commonExerciseIDs = append(commonExerciseIDs, id)
		}
	}

	//daily 카테고리 선별 정책 필요함!!
	var categoris []model.Category
	if err := service.db.Find(&categoris).Error; err != nil {
		return nil, errors.New("db error4")
	}

	categoryIds := []uint{}
	for _, c := range categoris {
		categoryIds = append(categoryIds, c.ID)
	}

	result := make(map[uint]RecomendResponse)
	if len(commonExerciseIDs) == 0 {
		log.Println("교집합이 없어서 최종적으로 할수있는 운동 없음")
		return nil, nil
	} else if len(commonExerciseIDs) <= RECOMMENDCOUNT {
		log.Println("운동 추천완료")
		var exercises []model.Exercise
		if err := service.db.Where("id IN ? ", commonExerciseIDs).Find(&exercises).Error; err != nil {
			return nil, errors.New("db error2")
		}

		for _, id := range categoryIds {
			for _, exercise := range exercises {
				if exercise.CategoryId == uint(id) {
					ex := ExerciseResponse{ID: exercise.ID, Name: exercise.Name}
					r := result[uint(id)]
					r.First = append(r.First, ex)
					result[uint(id)] = r
				}
			}
		}

	} else {
		log.Println("교집합이 많음")
		var exercises []model.Exercise
		var histories []model.History
		if err := service.db.Where("id IN ? ", commonExerciseIDs).Find(&exercises).Error; err != nil {
			return nil, errors.New("db error2")
		}
		if err := service.db.Where("exercise_id IN ? ", commonExerciseIDs).Find(&histories).Error; err != nil {
			return nil, errors.New("db error3")
		}
		exerciseCount := make(map[uint]int)
		for _, history := range histories {
			exerciseCount[history.ExerciseId]++
		}

		for _, id := range categoryIds {
			for _, exercise := range exercises {
				if exercise.CategoryId == uint(id) {
					log.Println("데일리 카테고리에 해당하는지 분류")
					ex := ExerciseResponse{ID: exercise.ID, Name: exercise.Name}
					r := result[uint(id)]
					r.First = append(r.First, ex)
					result[uint(id)] = r
				}
			}
		}

		//history 반영
		for key, value := range result {
			if len(value.First) > RECOMMENDCOUNT {
				log.Println("history 반영")
				sort.Slice(value.First, func(i, j int) bool {
					countI := exerciseCount[value.First[i].ID]
					countJ := exerciseCount[value.First[j].ID]
					return countI > countJ
				})
				value.First = value.First[:RECOMMENDCOUNT]
				value.Second = value.First[RECOMMENDCOUNT:]

				result[key] = value
			}
		}

	}

	return result, nil
}
