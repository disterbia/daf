// /user-service/service/service.go

package core

import (
	"coach-service/model"
	"errors"
	"log"
	"sort"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CoachService interface {
	login(request LoginRequest) (string, error)
	saveCategory(id uint, name string) (string, error)
	getCategoris() ([]CategoryResponse, error)
	saveExercise(id, categoryID uint, name string) (string, error)
	getMachines() ([]MachineDto, error)
	saveMachine(id uint, name string) (string, error)
	getPurposes() ([]PurposeDto, error)
	getRecommend(exerciseID uint) (RecommendResponse, error)
	getRecommends(page uint) ([]RecommendResponse, error)
	saveRecommend(request RecommendRequest) (string, error)
	searchRecommend(page uint, name string) ([]RecommendResponse, error)
}

type coachService struct {
	db *gorm.DB
}

func NewCoachService(db *gorm.DB) CoachService {
	return &coachService{db: db}
}

func (service *coachService) login(request LoginRequest) (string, error) {
	var u model.Admin
	if request.Password == "" {
		return "", errors.New("empty")
	}

	// 이메일로 사용자 조회
	if err := service.db.Where("email = ?", request.Email).First(&u).Error; err != nil {
		return "", errors.New("email not found")
	}

	// 비밀번호 비교
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(request.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// 새로운 JWT 토큰 생성
	tokenString, err := generateJWT(u)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (service *coachService) saveCategory(id uint, name string) (string, error) {
	result := service.db.Model(&model.Category{}).Where("id = ?", id).Update("name", name)
	if result.Error != nil {
		return "", errors.New("db error")
	}
	if result.RowsAffected == 0 {
		// 공백 제거한 name 생성
		trimmedName := strings.ReplaceAll(name, " ", "")
		// 기존에 동일한 name이 있는지 확인
		if err := service.db.Where("REPLACE(name, ' ', '') = ?", trimmedName).First(&model.Category{}).Error; err == nil {
			// 동일한 name이 존재하는 경우
			return "", errors.New("exist")
		} else if err != gorm.ErrRecordNotFound {
			// 데이터베이스 조회 중 오류가 발생한 경우
			return "", errors.New("db error2")
		}
		// 새로운 category 저장
		category := model.Category{Name: name}
		if err := service.db.Create(&category).Error; err != nil {
			return "", errors.New("db error3")
		}

		return "200", nil // 새로운 레코드 생성 성공
	}

	return "200", nil
}

func (service *coachService) getCategoris() ([]CategoryResponse, error) {
	var categoies []model.Category
	var categoyResponses []CategoryResponse
	if err := service.db.Preload("Exercises").Order("id DESC").Find(&categoies).Error; err != nil {
		return nil, errors.New("db error")
	}

	if err := copyStruct(categoies, &categoyResponses); err != nil {
		return nil, err
	}

	return categoyResponses, nil
}

func (service *coachService) saveExercise(id, categoryID uint, name string) (string, error) {
	result := service.db.Model(&model.Exercise{}).Where("id = ?", id).Update("name", name)
	if result.Error != nil {
		return "", errors.New("db error")
	}

	if result.RowsAffected == 0 {
		// 공백 제거한 name 생성
		trimmedName := strings.ReplaceAll(name, " ", "")

		// 기존에 동일한 name이 있는지 확인
		if err := service.db.Where("category_id = ? AND REPLACE(name, ' ', '') = ?", categoryID, trimmedName).First(&model.Exercise{}).Error; err == nil {
			// 동일한 name이 존재하는 경우
			return "", errors.New("exist")
		} else if err != gorm.ErrRecordNotFound {
			// 데이터베이스 조회 중 오류가 발생한 경우
			return "", errors.New("db error2")
		}

		// 새로운 exercise 저장
		exercise := model.Exercise{CategoryID: categoryID, Name: name}
		if err := service.db.Create(&exercise).Error; err != nil {
			return "", errors.New("db error3")
		}

		return "200", nil // 새로운 레코드 생성 성공
	}

	return "200", nil // 기존 레코드 업데이트 성공
}

func (service *coachService) getMachines() ([]MachineDto, error) {
	var machines []model.Machine
	var machineResponses []MachineDto
	if err := service.db.Find(&machines).Order("id DESC").Error; err != nil {
		return nil, errors.New("db error")
	}

	if err := copyStruct(machines, &machineResponses); err != nil {
		return nil, err
	}

	return machineResponses, nil
}

func (service *coachService) saveMachine(id uint, name string) (string, error) {
	result := service.db.Model(&model.Machine{}).Where("id = ?", id).Update("name", name)
	if result.Error != nil {
		return "", errors.New("db error")
	}

	if result.RowsAffected == 0 {
		// 공백 제거한 name 생성
		trimmedName := strings.ReplaceAll(name, " ", "")

		// 기존에 동일한 name이 있는지 확인
		if err := service.db.Where("REPLACE(name, ' ', '') = ?", trimmedName).First(&model.Machine{}).Error; err == nil {
			// 동일한 name이 존재하는 경우
			return "", errors.New("exist")
		} else if err != gorm.ErrRecordNotFound {
			// 데이터베이스 조회 중 오류가 발생한 경우
			return "", errors.New("db error2")
		}

		// 새로운 machine 저장
		machine := model.Machine{Name: name}
		if err := service.db.Create(&machine).Error; err != nil {
			return "", errors.New("db error3")
		}

		return "200", nil // 새로운 레코드 생성 성공
	}

	return "200", nil // 기존 레코드 업데이트 성공
}

func (service *coachService) getPurposes() ([]PurposeDto, error) {
	var purposes []model.Purpose
	var purposeResponse []PurposeDto
	if err := service.db.Find(&purposes).Order("id DESC").Error; err != nil {
		return nil, errors.New("db error")
	}

	if err := copyStruct(purposes, &purposeResponse); err != nil {
		return nil, err
	}

	return purposeResponse, nil
}

func (service *coachService) saveRecommend(request RecommendRequest) (string, error) {
	if err := validateRecommendRequest(request); err != nil {
		return "", err
	}
	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	// 추천운동
	if err := tx.Where("exercise_id = ?", request.ExerciseID).Unscoped().Delete(&model.Recommended{}).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error")
	}

	// 사용기구
	if err := tx.Where("exercise_id = ?", request.ExerciseID).Unscoped().Delete(&model.ExerciseMachine{}).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error3")
	}
	// 운동목적
	if err := tx.Where("exercise_id = ?", request.ExerciseID).Unscoped().Delete(&model.ExercisePurpose{}).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error5")
	}

	var recommends []model.Recommended
	recommends = append(recommends, model.Recommended{
		ExerciseID: request.ExerciseID,
		Asymmetric: request.Asymmetric,
		BodyFilter: request.BodyType,
		BodyTypeID: uint(TBODY),
		RomID:      request.TrRom,
	})
	recommends = append(recommends, model.Recommended{
		ExerciseID: request.ExerciseID,
		Asymmetric: request.Asymmetric,
		BodyFilter: request.BodyType,
		BodyTypeID: uint(LOCOBODY),
		RomID:      request.Locomotion,
	})

	log.Println(request.BodyRomClinicDegree)

	for bodyType, romClinicDegree := range request.BodyRomClinicDegree {
		if bodyType == uint(UBODY) || bodyType == uint(LBODY) {
			for rom, clinicDegree := range romClinicDegree {
				var checkClinic = make(map[uint]bool)
				for clinic, degree := range clinicDegree {
					if _, exists := checkClinic[clinic]; exists {
						continue // 중복된 clinic인 경우 처리하지 않음
					}
					recommends = append(recommends, model.Recommended{
						ExerciseID:        request.ExerciseID,
						Asymmetric:        request.Asymmetric,
						BodyFilter:        request.BodyType,
						BodyTypeID:        bodyType,
						RomID:             rom,
						ClinicalFeatureID: uintPointer(clinic),
						DegreeID:          uintPointer(degree),
					})
					checkClinic[clinic] = true
				}
			}
		}
	}
	if err := tx.Create(recommends).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error1")
	}

	// 추가 추천운동 생성
	var extras []model.Recommended
	degreePointer := uint(1)
	if request.BodyType == uint(UBODY) {
		for _, v := range CLINIC {
			extras = append(extras, model.Recommended{
				ExerciseID:        request.ExerciseID,
				Asymmetric:        request.Asymmetric,
				BodyFilter:        0,
				BodyTypeID:        uint(LBODY),
				RomID:             1,
				DegreeID:          &degreePointer,
				ClinicalFeatureID: uintPointer(uint(v)),
			})
		}
	} else if request.BodyType == uint(LBODY) {
		for _, v := range CLINIC {
			extras = append(extras, model.Recommended{
				ExerciseID:        request.ExerciseID,
				Asymmetric:        request.Asymmetric,
				BodyFilter:        0,
				BodyTypeID:        uint(UBODY),
				RomID:             1,
				DegreeID:          &degreePointer,
				ClinicalFeatureID: uintPointer(uint(v)),
			})
		}
	}

	if len(extras) > 0 {
		if err := tx.Create(&extras).Error; err != nil {
			tx.Rollback()
			return "", errors.New("db error: create extra recommendeds")
		}
	}

	var exerciseMachines []model.ExerciseMachine
	for _, id := range request.MachineIDs {
		exerciseMachines = append(exerciseMachines, model.ExerciseMachine{ExerciseID: request.ExerciseID, MachineID: id})
	}
	if err := tx.Create(exerciseMachines).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error4")
	}

	var exercisePurposes []model.ExercisePurpose
	for _, id := range request.PurposeIDs {
		exercisePurposes = append(exercisePurposes, model.ExercisePurpose{ExerciseID: request.ExerciseID, PurposeID: id})
	}
	if err := tx.Create(exercisePurposes).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error6")
	}
	tx.Commit()
	return "200", nil
}

func (service *coachService) getRecommend(exerciseID uint) (RecommendResponse, error) {
	var response RecommendResponse

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	// 추천운동 정보 가져오기
	var recommends []model.Recommended
	if err := tx.Where("exercise_id = ? AND body_filter != ? ", exerciseID, 0).Preload("Exercise.Category").Find(&recommends).Error; err != nil {
		tx.Rollback()
		return response, errors.New("db error")
	}
	if len(recommends) > 0 {
		response.Asymmetric = recommends[0].Asymmetric
		response.BodyRomClinicDegree = make(map[uint]map[uint]map[uint]uint)
		for _, rec := range recommends {
			response.Category = CategoryRequest{ID: rec.Exercise.CategoryID, Name: rec.Exercise.Category.Name}
			response.Exercise = ExerciseResponse{ID: rec.Exercise.ID, Name: rec.Exercise.Name, BodyType: rec.BodyTypeID}
			if rec.BodyTypeID == uint(TBODY) {
				response.TrRom = rec.RomID
				continue
			}
			if rec.BodyTypeID == uint(LOCOBODY) {
				response.Locomotion = rec.RomID
				continue
			}
			if response.BodyRomClinicDegree[rec.BodyTypeID] == nil {
				response.BodyRomClinicDegree[rec.BodyTypeID] = make(map[uint]map[uint]uint)
			}
			if response.BodyRomClinicDegree[rec.BodyTypeID][rec.RomID] == nil {
				response.BodyRomClinicDegree[rec.BodyTypeID][rec.RomID] = make(map[uint]uint)
			}
			response.BodyRomClinicDegree[rec.BodyTypeID][rec.RomID][*rec.ClinicalFeatureID] = *rec.DegreeID

		}
	}

	// 사용기구 정보 가져오기
	var exerciseMachines []model.ExerciseMachine
	if err := tx.Where("exercise_id = ?", exerciseID).Preload("Machine").Find(&exerciseMachines).Error; err != nil {
		tx.Rollback()
		return response, errors.New("db error")
	}
	response.Machines = make([]MachineDto, len(exerciseMachines))
	for i, machine := range exerciseMachines {
		response.Machines[i] = MachineDto{ID: machine.MachineID, Name: machine.Machine.Name}
	}

	// 운동목적 정보 가져오기
	var exercisePurposes []model.ExercisePurpose
	if err := tx.Where("exercise_id = ?", exerciseID).Preload("Purpose").Find(&exercisePurposes).Error; err != nil {
		tx.Rollback()
		return response, errors.New("db error")
	}
	response.Purposes = make([]PurposeDto, len(exercisePurposes))
	for i, purpose := range exercisePurposes {
		response.Purposes[i] = PurposeDto{ID: purpose.PurposeID, Name: purpose.Purpose.Name}
	}

	tx.Commit()
	return response, nil
}

func (service *coachService) getRecommends(page uint) ([]RecommendResponse, error) {
	var responses []RecommendResponse
	pageSize := 30
	offset := int(page) * pageSize

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	// 1. 전체 ExerciseID 목록 가져오기
	var exerciseIDs []uint
	if err := tx.Model(&model.Recommended{}).Distinct("exercise_id").Offset(offset).Limit(pageSize).Pluck("exercise_id", &exerciseIDs).Error; err != nil {
		tx.Rollback()
		return responses, errors.New("db error")
	}

	if len(exerciseIDs) == 0 {
		return responses, nil
	}

	// 2. 선택된 ExerciseID에 해당하는 추천 운동 데이터 가져오기
	var recommends []model.Recommended
	if err := tx.Where("exercise_id IN (?)", exerciseIDs).Where("body_filter != ?", 0).Order("id DESC").Preload("Exercise.Category").Find(&recommends).Error; err != nil {
		tx.Rollback()
		return responses, errors.New("db error")
	}

	// 사용기구 정보 가져오기
	var exerciseMachines []model.ExerciseMachine
	if err := tx.Where("exercise_id IN (?)", exerciseIDs).Preload("Machine").Find(&exerciseMachines).Error; err != nil {
		tx.Rollback()
		return responses, errors.New("db error")
	}

	// 운동목적 정보 가져오기
	var exercisePurposes []model.ExercisePurpose
	if err := tx.Where("exercise_id IN (?)", exerciseIDs).Preload("Purpose").Find(&exercisePurposes).Error; err != nil {
		tx.Rollback()
		return responses, errors.New("db error")
	}

	// 응답 구조체 생성
	exerciseIDToRecommend := make(map[uint]*RecommendResponse)
	for _, recommend := range recommends {
		if _, exists := exerciseIDToRecommend[recommend.ExerciseID]; !exists {
			exerciseIDToRecommend[recommend.ExerciseID] = &RecommendResponse{
				Category:            CategoryRequest{ID: recommend.Exercise.CategoryID, Name: recommend.Exercise.Category.Name},
				Exercise:            ExerciseResponse{ID: recommend.ExerciseID, Name: recommend.Exercise.Name, BodyType: recommend.BodyFilter},
				Asymmetric:          recommend.Asymmetric,
				BodyRomClinicDegree: make(map[uint]map[uint]map[uint]uint),
			}
		}
		if recommend.BodyTypeID == uint(TBODY) {
			exerciseIDToRecommend[recommend.ExerciseID].TrRom = recommend.RomID
			continue
		}
		if recommend.BodyTypeID == uint(LOCOBODY) {
			exerciseIDToRecommend[recommend.ExerciseID].Locomotion = recommend.RomID
			continue
		}
		if exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID] == nil {
			exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID] = make(map[uint]map[uint]uint)
		}
		if exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID][recommend.RomID] == nil {
			exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID][recommend.RomID] = make(map[uint]uint)
		}
		exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID][recommend.RomID][*recommend.ClinicalFeatureID] = *recommend.DegreeID

	}

	for _, machine := range exerciseMachines {
		if response, exists := exerciseIDToRecommend[machine.ExerciseID]; exists {
			response.Machines = append(response.Machines, MachineDto{ID: machine.MachineID, Name: machine.Machine.Name})
		}
	}

	for _, purpose := range exercisePurposes {
		if response, exists := exerciseIDToRecommend[purpose.ExerciseID]; exists {
			response.Purposes = append(response.Purposes, PurposeDto{ID: purpose.PurposeID, Name: purpose.Purpose.Name})
		}
	}

	// map을 slice로 변환하여 정렬
	var sortedResponses []*RecommendResponse
	for _, response := range exerciseIDToRecommend {
		sortedResponses = append(sortedResponses, response)
	}

	// 예시로 ExerciseID를 기준으로 정렬
	sort.Slice(sortedResponses, func(i, j int) bool {
		return sortedResponses[i].Exercise.ID < sortedResponses[j].Exercise.ID
	})

	// 최종 응답에 추가
	for _, response := range sortedResponses {
		responses = append(responses, *response)
	}

	tx.Commit()
	return responses, nil
}

func (service *coachService) searchRecommend(page uint, name string) ([]RecommendResponse, error) {
	log.Println("Original name:", name)
	var responses []RecommendResponse
	pageSize := 30
	offset := int(page) * pageSize

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	// 1. 전체 ExerciseID 목록 가져오기
	var exerciseIDs []uint
	chosungName := getChosung(name)
	log.Printf("Chosung name: %s", chosungName)
	if err := tx.Model(&model.Recommended{}).
		Joins("JOIN exercises ON exercises.id = recommendeds.exercise_id").
		Where("exercises.name LIKE ?", "%"+name+"%").
		Distinct("recommendeds.exercise_id").
		Offset(offset).Limit(pageSize).
		Pluck("recommendeds.exercise_id", &exerciseIDs).Error; err != nil {
		tx.Rollback()
		return responses, errors.New("db error")
	}

	if len(exerciseIDs) == 0 {
		return responses, nil
	}

	// 2. 선택된 ExerciseID에 해당하는 추천 운동 데이터 가져오기
	var recommends []model.Recommended
	if err := tx.Where("exercise_id IN (?)", exerciseIDs).
		Where("body_filter != ?", 0).
		Joins("JOIN exercises ON exercises.id = recommendeds.exercise_id").
		Order("CASE WHEN exercises.name LIKE '" + name + "%' THEN 0 ELSE 1 END, exercises.name, recommendeds.id DESC").
		Preload("Exercise.Category").
		Find(&recommends).Error; err != nil {
		tx.Rollback()
		return responses, errors.New("db error")
	}

	// 사용기구 정보 가져오기
	var exerciseMachines []model.ExerciseMachine
	if err := tx.Where("exercise_id IN (?)", exerciseIDs).Preload("Machine").Find(&exerciseMachines).Error; err != nil {
		tx.Rollback()
		return responses, errors.New("db error")
	}

	// 운동목적 정보 가져오기
	var exercisePurposes []model.ExercisePurpose
	if err := tx.Where("exercise_id IN (?)", exerciseIDs).Preload("Purpose").Find(&exercisePurposes).Error; err != nil {
		tx.Rollback()
		return responses, errors.New("db error")
	}

	// 응답 구조체 생성 및 초기화
	exerciseIDToRecommend := make(map[uint]*RecommendResponse)
	var sortedResponses []*RecommendResponse

	for _, recommend := range recommends {
		if _, exists := exerciseIDToRecommend[recommend.ExerciseID]; !exists {
			response := &RecommendResponse{
				Category:            CategoryRequest{ID: recommend.Exercise.CategoryID, Name: recommend.Exercise.Category.Name},
				Exercise:            ExerciseResponse{ID: recommend.ExerciseID, Name: recommend.Exercise.Name, BodyType: recommend.BodyFilter},
				Asymmetric:          recommend.Asymmetric,
				BodyRomClinicDegree: make(map[uint]map[uint]map[uint]uint),
			}
			exerciseIDToRecommend[recommend.ExerciseID] = response
			sortedResponses = append(sortedResponses, response)
		}

		if recommend.BodyTypeID == uint(TBODY) {
			exerciseIDToRecommend[recommend.ExerciseID].TrRom = recommend.RomID
			continue
		}
		if recommend.BodyTypeID == uint(LOCOBODY) {
			exerciseIDToRecommend[recommend.ExerciseID].Locomotion = recommend.RomID
			continue
		}
		if exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID] == nil {
			exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID] = make(map[uint]map[uint]uint)
		}
		if exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID][recommend.RomID] == nil {
			exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID][recommend.RomID] = make(map[uint]uint)
		}
		exerciseIDToRecommend[recommend.ExerciseID].BodyRomClinicDegree[recommend.BodyTypeID][recommend.RomID][*recommend.ClinicalFeatureID] = *recommend.DegreeID

	}

	for _, machine := range exerciseMachines {
		if response, exists := exerciseIDToRecommend[machine.ExerciseID]; exists {
			response.Machines = append(response.Machines, MachineDto{ID: machine.MachineID, Name: machine.Machine.Name})
		}
	}

	for _, purpose := range exercisePurposes {
		if response, exists := exerciseIDToRecommend[purpose.ExerciseID]; exists {
			response.Purposes = append(response.Purposes, PurposeDto{ID: purpose.PurposeID, Name: purpose.Purpose.Name})
		}
	}

	// 최종 응답에 추가
	for _, response := range sortedResponses {
		responses = append(responses, *response)
	}

	tx.Commit()
	return responses, nil
}
