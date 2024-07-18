// /user-service/service/service.go

package core

import (
	"coach-service/model"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CoachService interface {
	login(request LoginRequest) (string, error)
	saveCategory(id uint, name string) (string, error)
	getCategoris() ([]CategoryResponse, error)
	saveExercise(ExerciseRequest) (string, error)
	getMachines() ([]MachineDto, error)
	saveMachine(request MachineDto) (string, error)
	getPurposes() ([]PurposeDto, error)
	getRecommend(exerciseID uint) (RecommendResponse, error)
	// getRecommends(page uint) ([]RecommendResponse, error)
	saveRecommend(request RecommendRequest) (string, error)
	searchRecommend(page uint, name string) ([]RecommendResponse, error)
}

type coachService struct {
	db        *gorm.DB
	s3svc     *s3.S3
	bucket    string
	bucketUrl string
}

func NewCoachService(db *gorm.DB, s3svc *s3.S3, bucket string, bucketUrl string) CoachService {
	return &coachService{db: db, s3svc: s3svc, bucket: bucket, bucketUrl: bucketUrl}
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

	for _, v := range categoies {
		var exerciseResponses []ExerciseResponse
		for _, e := range v.Exercises {
			var explain []Explain
			if len(e.Explain) > 0 { // 설명 필드가 비어 있지 않은지 확인
				if err := json.Unmarshal(e.Explain, &explain); err != nil {
					return nil, err
				}
			}
			exerciseResponses = append(exerciseResponses, ExerciseResponse{ID: e.ID, Name: e.Name, Explain: explain})
		}
		categoyResponses = append(categoyResponses, CategoryResponse{ID: v.ID, Name: v.Name, Exercises: exerciseResponses})
	}

	return categoyResponses, nil
}

func (service *coachService) saveExercise(request ExerciseRequest) (string, error) {
	for i, v := range request.Explain {
		switch insertValue := v.Insert.(type) {
		case map[string]interface{}:
			if image, ok := insertValue["image"]; ok {
				imageString, ok := image.(string)
				if !ok {
					return "", errors.New("image field is not a string")
				}
				if !strings.HasPrefix(imageString, "http") {
					imgData, err := base64.StdEncoding.DecodeString(imageString)
					if err != nil {
						return "", err
					}

					// 이미지 포맷 체크
					contentType, ext, err := getImageFormat(imgData)
					if err != nil {
						return "", err
					}

					// 이미지 크기 조정 (10MB 제한)
					if len(imgData) > 10*1024*1024 {
						imgData, err = reduceImageSize(imgData)
						if err != nil {
							return "", err
						}
					}

					// S3에 이미지 및 썸네일 업로드
					url, err := uploadImagesToS3(imgData, contentType, ext, service.s3svc, service.bucket, service.bucketUrl, strconv.FormatUint(uint64(request.ID), 10))
					if err != nil {
						return "", err
					}
					request.Explain[i].Insert = map[string]interface{}{"image": url}
				}
			}
		case string:
			// v.Insert is a string, nothing to do
		default:
			return "", errors.New("unexpected insert value type")
		}
	}

	// QuillJson을 JSON 문자열로 변환
	explainJson, err := json.Marshal(request.Explain)
	if err != nil {
		return "", err
	}
	updates := map[string]interface{}{
		"name":    request.Name,
		"explain": explainJson, // 여기에 업데이트할 다른 필드를 추가
	}

	result := service.db.Model(&model.Exercise{}).Where("id = ?", request.ID).Updates(updates)
	if result.Error != nil {
		return "", errors.New("db error")
	}

	if result.RowsAffected == 0 {
		// 공백 제거한 name 생성
		trimmedName := strings.ReplaceAll(request.Name, " ", "")

		// 기존에 동일한 name이 있는지 확인
		if err := service.db.Where("category_id = ? AND REPLACE(name, ' ', '') = ?", request.CategoryId, trimmedName).First(&model.Exercise{}).Error; err == nil {
			// 동일한 name이 존재하는 경우
			return "", errors.New("exist")
		} else if err != gorm.ErrRecordNotFound {
			// 데이터베이스 조회 중 오류가 발생한 경우
			return "", errors.New("db error2")
		}

		// 새로운 exercise 저장
		exercise := model.Exercise{CategoryID: request.CategoryId, Name: request.Name, Explain: explainJson}
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

func (service *coachService) saveMachine(request MachineDto) (string, error) {
	updates := map[string]interface{}{
		"name":         request.Name,
		"machine_type": request.MachineType,
		"memo":         request.Memo,
	}
	result := service.db.Model(&model.Machine{}).Where("id = ?", request.ID).Updates(updates)
	if result.Error != nil {
		return "", errors.New("db error")
	}

	if result.RowsAffected == 0 {
		// 공백 제거한 name 생성
		trimmedName := strings.ReplaceAll(request.Name, " ", "")

		// 기존에 동일한 name이 있는지 확인
		if err := service.db.Where("REPLACE(name, ' ', '') = ?", trimmedName).First(&model.Machine{}).Error; err == nil {
			// 동일한 name이 존재하는 경우
			return "", errors.New("exist")
		} else if err != gorm.ErrRecordNotFound {
			// 데이터베이스 조회 중 오류가 발생한 경우
			return "", errors.New("db error2")
		}

		// 새로운 machine 저장
		machine := model.Machine{Name: request.Name, MachineType: request.MachineType, Memo: request.Memo}
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

	// if err := tx.Where("recommended_id = ?", request.ExerciseID).Unscoped().Delete(&model.RecommendedClinicalDegree{}).Error; err != nil {
	// 	tx.Rollback()
	// 	return "", errors.New("db error")
	// }

	// if err := tx.Where("recommended_id = ?", request.ExerciseID).Unscoped().Delete(&model.RecommendedJointRom{}).Error; err != nil {
	// 	tx.Rollback()
	// 	return "", errors.New("db error")
	// }

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

	// 측정항목
	if err := tx.Where("exercise_id = ?", request.ExerciseID).Unscoped().Delete(&model.ExerciseMeasure{}).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error5")
	}

	recommend := model.Recommended{
		ExerciseID:   request.ExerciseID,
		IsAsymmetric: request.IsAsymmetric,
		BodyTypeID:   request.BodyType,
		TRomID:       request.TrRom,
		LocoRomID:    request.Locomotion,
	}

	if err := tx.Create(&recommend).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error6")
	}

	var rcds []model.RecommendedClinicalDegree
	var rjrs []model.RecommendedJointRom

	for i, v := range request.Afcs {
		rjrs = append(rjrs, model.RecommendedJointRom{RecommendedID: recommend.ID, JointActionID: v.JointAction, RomID: v.Rom})
		// 반대부위 추천운동 생성
		if i == 1 {
			if request.BodyType == uint(UBODY) {
				rjrs = append(rjrs, model.RecommendedJointRom{RecommendedID: recommend.ID, JointActionID: uint(HIP), RomID: 1})
				rjrs = append(rjrs, model.RecommendedJointRom{RecommendedID: recommend.ID, JointActionID: uint(KNEE), RomID: 1})
			} else if request.BodyType == uint(LBODY) {
				rjrs = append(rjrs, model.RecommendedJointRom{RecommendedID: recommend.ID, JointActionID: uint(SHOULDER), RomID: 1})
				rjrs = append(rjrs, model.RecommendedJointRom{RecommendedID: recommend.ID, JointActionID: uint(ELBOW), RomID: 1})
			}
		}

		for clinic, degree := range v.ClinicDegree {
			rcds = append(rcds, model.RecommendedClinicalDegree{RecommendedID: recommend.ID, JointActionID: v.JointAction, ClinicalFeatureID: clinic, DegreeID: degree})
			// 반대부위 추천운동 생성
			if i == 1 {
				if request.BodyType == uint(UBODY) {
					rcds = append(rcds, model.RecommendedClinicalDegree{RecommendedID: recommend.ID, JointActionID: uint(HIP), ClinicalFeatureID: clinic, DegreeID: 1})
					rcds = append(rcds, model.RecommendedClinicalDegree{RecommendedID: recommend.ID, JointActionID: uint(KNEE), ClinicalFeatureID: clinic, DegreeID: 1})
				} else if request.BodyType == uint(LBODY) {
					rcds = append(rcds, model.RecommendedClinicalDegree{RecommendedID: recommend.ID, JointActionID: uint(SHOULDER), ClinicalFeatureID: clinic, DegreeID: 1})
					rcds = append(rcds, model.RecommendedClinicalDegree{RecommendedID: recommend.ID, JointActionID: uint(ELBOW), ClinicalFeatureID: clinic, DegreeID: 1})
				}
			}
		}
	}

	if err := tx.Create(&rjrs).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error6")
	}
	if err := tx.Create(&rcds).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error7")
	}

	var exerciseMachines []model.ExerciseMachine
	for _, id := range request.MachineIDs {
		exerciseMachines = append(exerciseMachines, model.ExerciseMachine{ExerciseID: request.ExerciseID, MachineID: id})
	}
	if err := tx.Create(&exerciseMachines).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error4")
	}

	var exercisePurposes []model.ExercisePurpose
	for _, id := range request.PurposeIDs {
		exercisePurposes = append(exercisePurposes, model.ExercisePurpose{ExerciseID: request.ExerciseID, PurposeID: id})
	}
	if err := tx.Create(&exercisePurposes).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error6")
	}

	var exerciseMeasure []model.ExerciseMeasure
	for _, id := range request.MeasureIds {
		exerciseMeasure = append(exerciseMeasure, model.ExerciseMeasure{ExerciseID: request.ExerciseID, MeasureID: id})
	}

	if len(exerciseMeasure) != 0 {
		if err := tx.Create(&exerciseMeasure).Error; err != nil {
			tx.Rollback()
			return "", errors.New("db error7")
		}
	}

	tx.Commit()
	return "200", nil
}

func (service *coachService) getRecommend(exerciseID uint) (RecommendResponse, error) {
	var response RecommendResponse
	// 추천운동 정보 가져오기
	var recommend model.Recommended
	if err := service.db.Where("exercise_id = ?", exerciseID).Preload("Exercise.Category").Preload("ClinicalDegrees").Preload("JointRoms").First(&recommend).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return response, errors.New("db error")
		} else {
			return response, errors.New("not found")
		}
	}

	var explain []Explain
	if len(recommend.Exercise.Explain) > 0 {
		if err := json.Unmarshal(recommend.Exercise.Explain, &explain); err != nil {
			return RecommendResponse{}, err
		}
	}
	var afcs []RecommendAfc

	jointClinicDegree := make(map[uint]map[uint]uint)
	for _, w := range recommend.ClinicalDegrees {
		if jointClinicDegree[w.JointActionID] == nil {
			jointClinicDegree[w.JointActionID] = make(map[uint]uint)
		}
		jointClinicDegree[w.JointActionID][w.ClinicalFeatureID] = w.DegreeID
	}

	for _, v := range recommend.JointRoms {
		if recommend.BodyTypeID == uint(UBODY) {
			if v.JointActionID == uint(SHOULDER) || v.JointActionID == uint(ELBOW) {
				afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
			}
		} else if recommend.BodyTypeID == uint(LBODY) {
			if v.JointActionID == uint(HIP) || v.JointActionID == uint(KNEE) {
				afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
			}
		} else {
			afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
		}
	}

	response = RecommendResponse{Category: CategoryRequest{ID: recommend.Exercise.CategoryID, Name: recommend.Exercise.Category.Name},
		Exercise:     ExerciseResponse{ID: recommend.Exercise.ID, Name: recommend.Exercise.Name, Explain: explain},
		IsAsymmetric: recommend.IsAsymmetric,
		TrRom:        recommend.TRomID,
		Locomotion:   recommend.LocoRomID,
		BodyType:     recommend.BodyTypeID,
		Afcs:         afcs}

	// 사용기구 정보 가져오기
	var exerciseMachines []model.ExerciseMachine
	if err := service.db.Where("exercise_id = ?", exerciseID).Preload("Machine").Find(&exerciseMachines).Error; err != nil {
		return response, errors.New("db error1")
	}
	response.Machines = make([]MachineDto, len(exerciseMachines))
	for i, machine := range exerciseMachines {
		response.Machines[i] = MachineDto{ID: machine.MachineID, Name: machine.Machine.Name}
	}

	// 운동목적 정보 가져오기
	var exercisePurposes []model.ExercisePurpose
	if err := service.db.Where("exercise_id = ?", exerciseID).Preload("Purpose").Find(&exercisePurposes).Error; err != nil {
		return response, errors.New("db error2")
	}
	response.Purposes = make([]PurposeDto, len(exercisePurposes))
	for i, purpose := range exercisePurposes {
		response.Purposes[i] = PurposeDto{ID: purpose.PurposeID, Name: purpose.Purpose.Name}
	}

	// 측정항목 정보 가져오기
	var exerciseMeasure []model.ExerciseMeasure
	if err := service.db.Where("exercise_id = ?", exerciseID).Preload("Measure").Find(&exerciseMeasure).Error; err != nil {
		return response, errors.New("db error3")
	}
	response.Measures = make([]MeasureDto, len(exerciseMeasure))
	for i, measure := range exerciseMeasure {
		response.Measures[i] = MeasureDto{ID: measure.MeasureID, Name: measure.Measure.Name}
	}

	return response, nil
}

// func (service *coachService) getRecommends(page uint) ([]RecommendResponse, error) {
// 	var responses []RecommendResponse
// 	pageSize := 30
// 	offset := int(page) * pageSize

// 	// 1. 전체 ExerciseID 목록 가져오기
// 	var exerciseIDs []uint
// 	if err := service.db.Model(&model.Recommended{}).Distinct("exercise_id").Offset(offset).Limit(pageSize).Pluck("exercise_id", &exerciseIDs).Error; err != nil {
// 		return nil, errors.New("db error")
// 	}

// 	if len(exerciseIDs) == 0 {
// 		return nil, nil
// 	}

// 	// 2. 선택된 ExerciseID에 해당하는 추천 운동 데이터 가져오기
// 	var recommends []model.Recommended
// 	if err := service.db.Where("exercise_id IN (?)", exerciseIDs).Order("id DESC").Preload("Exercise.Category").Preload("ClinicalDegrees").Preload("JointRoms").Find(&recommends).Error; err != nil {
// 		return nil, errors.New("db error1")
// 	}

// 	// 사용기구 정보 가져오기
// 	var exerciseMachines []model.ExerciseMachine
// 	if err := service.db.Where("exercise_id IN (?)", exerciseIDs).Preload("Machine").Find(&exerciseMachines).Error; err != nil {
// 		return nil, errors.New("db error2")
// 	}

// 	// 운동목적 정보 가져오기
// 	var exercisePurposes []model.ExercisePurpose
// 	if err := service.db.Where("exercise_id IN (?)", exerciseIDs).Preload("Purpose").Find(&exercisePurposes).Error; err != nil {
// 		return nil, errors.New("db error3")
// 	}

// 	// 운동목적 정보 가져오기
// 	var exerciseMeasures []model.ExerciseMeasure
// 	if err := service.db.Where("exercise_id IN (?)", exerciseIDs).Preload("Measure").Find(&exerciseMeasures).Error; err != nil {
// 		return nil, errors.New("db error4")
// 	}

// 	// 응답 구조체 생성
// 	exerciseIDToRecommend := make(map[uint]*RecommendResponse)
// 	var sortedResponses []*RecommendResponse
// 	for _, recommend := range recommends {
// 		var explain []Explain
// 		if len(recommend.Exercise.Explain) > 0 {
// 			if err := json.Unmarshal(recommend.Exercise.Explain, &explain); err != nil {
// 				return nil, err
// 			}
// 		}
// 		if _, exists := exerciseIDToRecommend[recommend.ExerciseID]; !exists {
// 			response := &RecommendResponse{
// 				Category:     CategoryRequest{ID: recommend.Exercise.CategoryID, Name: recommend.Exercise.Category.Name},
// 				Exercise:     ExerciseResponse{ID: recommend.ExerciseID, Name: recommend.Exercise.Name, Explain: explain},
// 				IsAsymmetric: recommend.IsAsymmetric,
// 				TrRom:        recommend.TRomID,
// 				Locomotion:   recommend.LocoRomID,
// 				BodyType:     recommend.BodyTypeID,
// 			}
// 			exerciseIDToRecommend[recommend.ExerciseID] = response
// 			sortedResponses = append(sortedResponses, response)
// 		}

// 		var afcs []RecommendAfc

// 		jointClinicDegree := make(map[uint]map[uint]uint)
// 		for _, w := range recommend.ClinicalDegrees {
// 			if jointClinicDegree[w.JointActionID] == nil {
// 				jointClinicDegree[w.JointActionID] = make(map[uint]uint)
// 			}
// 			jointClinicDegree[w.JointActionID][w.ClinicalFeatureID] = w.DegreeID
// 		}

// 		for _, v := range recommend.JointRoms {
// 			if recommend.BodyTypeID == uint(UBODY) {
// 				if v.JointActionID == uint(SHOULDER) || v.JointActionID == uint(ELBOW) {
// 					afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
// 				}
// 			} else if recommend.BodyTypeID == uint(LBODY) {
// 				if v.JointActionID == uint(HIP) || v.JointActionID == uint(KNEE) {
// 					afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
// 				}
// 			} else {
// 				afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
// 			}

// 		}
// 		exerciseIDToRecommend[recommend.ExerciseID].Afcs = afcs

// 	}

// 	for _, machine := range exerciseMachines {
// 		if recommend, exists := exerciseIDToRecommend[machine.ExerciseID]; exists {
// 			recommend.Machines = append(recommend.Machines, MachineDto{ID: machine.MachineID, Name: machine.Machine.Name})
// 		}
// 	}

// 	for _, purpose := range exercisePurposes {
// 		if recommend, exists := exerciseIDToRecommend[purpose.ExerciseID]; exists {
// 			recommend.Purposes = append(recommend.Purposes, PurposeDto{ID: purpose.PurposeID, Name: purpose.Purpose.Name})
// 		}
// 	}

// 	for _, measure := range exerciseMeasures {
// 		if recommend, exists := exerciseIDToRecommend[measure.ExerciseID]; exists {
// 			recommend.Measures = append(recommend.Measures, MeasureDto{ID: measure.MeasureID, Name: measure.Measure.Name})
// 		}
// 	}

// 	// 최종 응답에 추가
// 	for _, response := range sortedResponses {
// 		responses = append(responses, *response)
// 	}

// 	return responses, nil
// }

func (service *coachService) searchRecommend(page uint, name string) ([]RecommendResponse, error) {
	log.Println("Original name:", name)
	var responses []RecommendResponse
	pageSize := 30
	offset := int(page) * pageSize

	// 1. 전체 ExerciseID 목록 가져오기
	var exerciseIDs []uint
	if err := service.db.Model(&model.Recommended{}).
		Joins("JOIN exercises ON exercises.id = recommendeds.exercise_id").
		Where("exercises.name LIKE ?", "%"+name+"%").
		Distinct("recommendeds.exercise_id").
		Offset(offset).Limit(pageSize).
		Pluck("recommendeds.exercise_id", &exerciseIDs).Error; err != nil {
		return responses, errors.New("db error")
	}

	if len(exerciseIDs) == 0 {
		return responses, nil
	}

	// 2. 선택된 ExerciseID에 해당하는 추천 운동 데이터 가져오기
	var recommends []model.Recommended
	if err := service.db.Where("exercise_id IN (?)", exerciseIDs).
		Joins("JOIN exercises ON exercises.id = recommendeds.exercise_id").
		Preload("Exercise.Category").Preload("ClinicalDegrees").Preload("JointRoms").
		Order("CASE WHEN exercises.name LIKE '" + name + "%' THEN 0 ELSE 1 END, exercises.name, recommendeds.id DESC").
		Find(&recommends).Error; err != nil {
		return responses, errors.New("db error1")
	}

	// 사용기구 정보 가져오기
	var exerciseMachines []model.ExerciseMachine
	if err := service.db.Where("exercise_id IN (?)", exerciseIDs).Preload("Machine").Find(&exerciseMachines).Error; err != nil {
		return responses, errors.New("db error2")
	}

	// 운동목적 정보 가져오기
	var exercisePurposes []model.ExercisePurpose
	if err := service.db.Where("exercise_id IN (?)", exerciseIDs).Preload("Purpose").Find(&exercisePurposes).Error; err != nil {
		return responses, errors.New("db error3")
	}

	// 측정항목 정보 가져오기
	var exerciseMeasures []model.ExerciseMeasure
	if err := service.db.Where("exercise_id IN (?)", exerciseIDs).Preload("Measure").Find(&exerciseMeasures).Error; err != nil {
		return responses, errors.New("db error4")
	}

	// 응답 구조체 생성 및 초기화
	exerciseIDToRecommend := make(map[uint]*RecommendResponse)
	var sortedResponses []*RecommendResponse

	for _, recommend := range recommends {
		var explain []Explain
		if len(recommend.Exercise.Explain) > 0 {
			if err := json.Unmarshal(recommend.Exercise.Explain, &explain); err != nil {
				return nil, err
			}
		}
		if _, exists := exerciseIDToRecommend[recommend.ExerciseID]; !exists {
			response := &RecommendResponse{
				Category:     CategoryRequest{ID: recommend.Exercise.CategoryID, Name: recommend.Exercise.Category.Name},
				Exercise:     ExerciseResponse{ID: recommend.ExerciseID, Name: recommend.Exercise.Name, Explain: explain},
				IsAsymmetric: recommend.IsAsymmetric,
				TrRom:        recommend.TRomID,
				Locomotion:   recommend.LocoRomID,
				BodyType:     recommend.BodyTypeID,
			}
			exerciseIDToRecommend[recommend.ExerciseID] = response
			sortedResponses = append(sortedResponses, response)
		}

		var afcs []RecommendAfc

		jointClinicDegree := make(map[uint]map[uint]uint)
		for _, w := range recommend.ClinicalDegrees {
			if jointClinicDegree[w.JointActionID] == nil {
				jointClinicDegree[w.JointActionID] = make(map[uint]uint)
			}
			jointClinicDegree[w.JointActionID][w.ClinicalFeatureID] = w.DegreeID
		}

		for _, v := range recommend.JointRoms {
			if recommend.BodyTypeID == uint(UBODY) {
				if v.JointActionID == uint(SHOULDER) || v.JointActionID == uint(ELBOW) {
					afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
				}
			} else if recommend.BodyTypeID == uint(LBODY) {
				if v.JointActionID == uint(HIP) || v.JointActionID == uint(KNEE) {
					afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
				}
			} else {
				afcs = append(afcs, RecommendAfc{JointAction: v.JointActionID, Rom: v.RomID, ClinicDegree: jointClinicDegree[v.JointActionID]})
			}
		}
		exerciseIDToRecommend[recommend.ExerciseID].Afcs = afcs

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

	for _, measure := range exerciseMeasures {
		if response, exists := exerciseIDToRecommend[measure.ExerciseID]; exists {
			response.Measures = append(response.Measures, MeasureDto{ID: measure.MeasureID, Name: measure.Measure.Name})
		}
	}

	// 최종 응답에 추가
	for _, response := range sortedResponses {
		responses = append(responses, *response)
	}

	return responses, nil
}
