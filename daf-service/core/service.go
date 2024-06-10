// /user-service/service/service.go

package core

import (
	"daf-service/model"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DafService interface {
	setUser(userRequest UserJointActionRequest) (string, error)
	getUser(id uint) (UserJointActionResponse, error)
	getRecommend(id uint) ([]ExerciseResponse, error)
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

func (service *dafService) getRecommend(id uint) ([]ExerciseResponse, error) {

	var userJointActions []model.UserJointAction
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

		// BodyCompositionId를 키로 사용하는 그룹에 데이터 추가
		bodyCompId := userJointAction.JointAction.BodyCompositionId
		data := groupData[bodyCompId]
		data.romList += userJointAction.RomId
		data.clinicList = append(data.clinicList, userJointAction.ClinicalFeature.ID)
		data.degreeList += userJointAction.DegreeId
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
			tr = SearchData{bodyType: uint(TBODY), rom: romAver, clinic: clinicAver, degree: degreeAver}
		case uint(LOCO):
			loco = SearchData{bodyType: uint(LOCOBODY), rom: romAver, clinic: clinicAver, degree: degreeAver}
		}

		fmt.Printf("BodyCompositionId: %d, ROM 평균: %d, Degree 평균: %d, Clinic 평균: %d\n", bodyCompId, romAver, degreeAver, clinicAver)
	}

	// var recommends []model.Recommended

	var recommends []model.Recommended
	var result RecomendResponse
	err := service.db.Where(`
	(body_type_id = ? AND rom_id <= ?) OR
	(body_type_id = ? AND rom_id <= ?) OR
	(body_type_id = ? AND clinical_feature_id = ? AND rom_id <= ? AND degree_id <= ?) OR
	(body_type_id = ? AND clinical_feature_id = ? AND rom_id <= ? AND degree_id <= ?) OR
	(body_type_id = ? AND clinical_feature_id = ? AND rom_id <= ? AND degree_id <= ?) OR
	(body_type_id = ? AND clinical_feature_id = ? AND rom_id <= ? AND degree_id <= ?)`,
		tr.bodyType, tr.rom,
		loco.bodyType, loco.rom,
		ulav.bodyType, ulav.clinic, ulav.rom, ulav.degree,
		urav.bodyType, urav.clinic, urav.rom, urav.degree,
		llav.bodyType, llav.clinic, llav.rom, llav.degree,
		lrav.bodyType, lrav.clinic, lrav.rom, lrav.degree).
		Find(&recommends).Error
	if err != nil {
		return nil, errors.New("db error")
	}

	var recommendsTR, recommendsLoco, recommendsUlav, recommendsUrav, recommendsLlav, recommendsLrav []model.Recommended

	for _, rec := range recommends {
		switch {
		case rec.BodyTypeID == tr.bodyType && rec.RomID <= tr.rom:
			recommendsTR = append(recommendsTR, rec)
		case rec.BodyTypeID == loco.bodyType && rec.RomID <= loco.rom:
			recommendsLoco = append(recommendsLoco, rec)
		case rec.BodyTypeID == ulav.bodyType && rec.ClinicalFeatureID == ulav.clinic && rec.RomID <= ulav.rom && rec.DegreeID <= ulav.degree:
			recommendsUlav = append(recommendsUlav, rec)
		case rec.BodyTypeID == urav.bodyType && rec.ClinicalFeatureID == urav.clinic && rec.RomID <= urav.rom && rec.DegreeID <= urav.degree:
			recommendsUrav = append(recommendsUrav, rec)
		case rec.BodyTypeID == llav.bodyType && rec.ClinicalFeatureID == llav.clinic && rec.RomID <= llav.rom && rec.DegreeID <= llav.degree:
			recommendsLlav = append(recommendsLlav, rec)
		case rec.BodyTypeID == lrav.bodyType && rec.ClinicalFeatureID == lrav.clinic && rec.RomID <= lrav.rom && rec.DegreeID <= lrav.degree:
			recommendsLrav = append(recommendsLrav, rec)
		}
	}

	exerciseIDMap := make(map[uint]int)

	countExerciseID := func(recommends []model.Recommended) {
		for _, rec := range recommends {
			exerciseIDMap[rec.ExerciseID]++
		}
	}

	countExerciseID(recommendsTR)
	countExerciseID(recommendsLoco)
	countExerciseID(recommendsUlav)
	countExerciseID(recommendsUrav)
	countExerciseID(recommendsLlav)
	countExerciseID(recommendsLrav)

	var commonExerciseIDs []uint
	for id, count := range exerciseIDMap {
		if count == 6 { // 모든 슬라이스에 포함된 ExerciseID
			commonExerciseIDs = append(commonExerciseIDs, id)
		}
	}

	sqlQuery := getQuery(1)
	log.Println("최초 쿼리실행")
	if err := service.db.Raw(sqlQuery,
		tr.bodyType, tr.rom,
		ulav.bodyType, ulav.clinic, ulav.rom, ulav.degree,
		urav.bodyType, urav.clinic, urav.rom, urav.degree,
		llav.bodyType, llav.clinic, llav.rom, llav.degree,
		lrav.bodyType, lrav.clinic, lrav.rom, lrav.degree).
		Scan(&recommends).Error; err != nil {
		return nil, errors.New("db error")
	}

	if len(recommends) > DafCount {
		log.Println("추천운동이 많음")
		sqlQuery := getQuery(2)
		log.Println("부위별로 rom이 이하가 아닌 rom이 같은 운동 조회")
		if err := service.db.Raw(sqlQuery,
			tr.bodyType, tr.rom,
			ulav.bodyType, ulav.clinic, ulav.rom, ulav.degree,
			urav.bodyType, urav.clinic, urav.rom, urav.degree,
			llav.bodyType, llav.clinic, llav.rom, llav.degree,
			lrav.bodyType, lrav.clinic, lrav.rom, lrav.degree).
			Scan(&recommends).Error; err != nil {
			return nil, errors.New("db error")
		}

		if len(recommends) > DafCount {
			log.Println("rom을 정확히해도 추천운동이 많음")
			sqlQuery = getQuery(3)
			log.Println("부위별로 rom도 같고 degree도 같은운동 조회")
			if err := service.db.Raw(sqlQuery,
				tr.bodyType, tr.rom,
				ulav.bodyType, ulav.clinic, ulav.rom, ulav.degree,
				urav.bodyType, urav.clinic, urav.rom, urav.degree,
				llav.bodyType, llav.clinic, llav.rom, llav.degree,
				lrav.bodyType, lrav.clinic, lrav.rom, lrav.degree).
				Scan(&recommends).Error; err != nil {
				return nil, errors.New("db error")
			}
		} else if len(recommends) == 0 {
			log.Println("rom을 정확히하니 추천운동이 없음")

		} else {
			log.Println("rom을 정확히하니 개수가 작음")

		}

		if len(recommends) > DafCount {
			log.Println("rom과 degree를 정확히해도 추천운동이 많음")
			// 랜덤 생성기 생성
			source := rand.NewSource(time.Now().UnixNano())
			rng := rand.New(source)

			// 슬라이스를 랜덤하게 섞기
			rng.Shuffle(len(recommends), func(i, j int) {
				recommends[i], recommends[j] = recommends[j], recommends[i]
			})

			// 슬라이스의 처음 5개 요소를 가져오기
			recommends = recommends[:5]
			for _, v := range recommends {
				temp := ExerciseResponse{ID: v.ExerciseID, Name: v.Exercise.Name, Category: CategoryResponse{ID: v.Exercise.CategoryId, Name: v.Exercise.Category.Name}}
				result.First = append(result.First, temp)
			}

		} else if len(recommends) == 0 {
			log.Println("rom과 degree를 정확히하니 추천운동이 없음")
		} else {
			log.Println("rom과 degree를 정확히하니 추천운동이 작음")
		}

	} else if len(recommends) == 0 {
		log.Println("처음부터 추천운동이 없음")
	} else {
		log.Println("추천운동이 작음")
	}

	// Initial ROM values
	roms := []uint{tr.rom, ulav.rom, urav.rom, llav.rom, lrav.rom}
	conditions := []SearchData{tr, ulav, urav, llav, lrav}

	// Function to check if a specific condition is met
	checkCondition := func(recommends []model.Recommended, bodyTypeID, clinicalFeatureID, romID, degreeID uint) bool {
		for _, r := range recommends {
			if r.BodyTypeID == bodyTypeID && r.ClinicalFeatureID == clinicalFeatureID && r.RomID <= romID && r.DegreeID <= degreeID {
				return true
			}
		}
		return false
	}

	// Loop until all conditions are met
	for {
		allConditionsMet := true
		for i, cond := range conditions {
			if !checkCondition(recommends, cond.bodyType, cond.clinic, roms[i], cond.degree) {
				roms[i]++
				allConditionsMet = false
			}
		}

		if allConditionsMet {
			break
		}
	}

	// 운동들의 등장 횟수를 세기 위한 맵
	exerciseFrequency := make(map[uint]map[uint]int) // categoryID -> exerciseID -> count
	for _, recommend := range recommends {
		categoryID := recommend.Exercise.CategoryId
		exerciseID := recommend.Exercise.ID

		if _, exists := exerciseFrequency[categoryID]; !exists {
			exerciseFrequency[categoryID] = make(map[uint]int)
		}
		exerciseFrequency[categoryID][exerciseID]++
	}

	type ExerciseRank struct {
		ID       uint
		Name     string
		Category CategoryResponse
		Count    int
	}

	var rankedExercises []ExerciseRank
	for categoryID, exercises := range exerciseFrequency {
		for exerciseID, count := range exercises {
			for _, recommend := range recommends {
				if recommend.Exercise.ID == exerciseID {
					rankedExercises = append(rankedExercises, ExerciseRank{
						ID:       exerciseID,
						Name:     recommend.Exercise.Name,
						Category: CategoryResponse{ID: categoryID, Name: recommend.Exercise.Category.Name},
						Count:    count,
					})
				}
			}
		}
	}

	// 등장 횟수 기준으로 정렬
	sort.Slice(rankedExercises, func(i, j int) bool {
		return rankedExercises[i].Count > rankedExercises[j].Count
	})

	// 상위 3개의 운동만 선택
	if len(rankedExercises) > 3 {
		rankedExercises = rankedExercises[:3]
	}

	var response []ExerciseResponse
	for _, rank := range rankedExercises {
		response = append(response, ExerciseResponse{
			ID:       rank.ID,
			Name:     rank.Name,
			Category: rank.Category,
		})
	}

	return response, nil
}
