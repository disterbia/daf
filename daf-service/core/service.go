// /user-service/service/service.go

package core

import (
	"daf-service/model"
	"errors"
	"fmt"
	"log"
	"sort"

	"gorm.io/gorm"
)

type DafService interface {
	getRecommends(id uint) (map[uint]RecomendResponse, error)
}

type dafService struct {
	db *gorm.DB
}

func NewDafService(db *gorm.DB) DafService {
	return &dafService{db: db}
}

func (service *dafService) getRecommends(id uint) (map[uint]RecomendResponse, error) {

	//유저 부위별 코드 가져오기
	var userAfcs []model.UserAfc
	if err := service.db.Where("uid = ?", id).Find(&userAfcs).Error; err != nil {
		return nil, errors.New("db error")
	}

	type GroupData struct {
		romList         []uint
		clinicList      []uint
		degreeList      []uint
		amputationCount int
	}
	groupData := make(map[uint]GroupData)

	type SearchData struct {
		bodyType   uint
		rom        uint
		clinic     uint
		degree     uint
		amputation int
	}
	var ulav, urav, llav, lrav, tr, loco SearchData
	for _, userAfc := range userAfcs {
		// BodyCompositionID를 키로 사용하는 그룹에 데이터 추가
		bodyCompId := userAfc.BodyCompositionID
		data := groupData[bodyCompId]
		if userAfc.RomID != nil {
			data.romList = append(data.romList, *userAfc.RomID)
		}
		if userAfc.ClinicalFeatureID != nil {
			data.clinicList = append(data.clinicList, *userAfc.ClinicalFeatureID)
			if *userAfc.ClinicalFeatureID == uint(AC) {
				data.amputationCount += 1 ///절단 등장 횟수
			}
		}
		if userAfc.DegreeID != nil {
			data.degreeList = append(data.degreeList, *userAfc.DegreeID)
		}
		groupData[bodyCompId] = data
	}

	// 그룹별 평균 계산
	for bodyCompId, data := range groupData {
		romAver := sum(data.romList) / uint(len(data.romList))
		degreeAver := sum(data.degreeList) / uint(len(data.degreeList))
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
		amputation := 5
		amputation = amputation - data.amputationCount
		if amputation == 5 {
			amputation = 0
		}
		switch bodyCompId {
		// responseValue
		case uint(UL):
			ulav = SearchData{bodyType: uint(UBODY), rom: romAver, clinic: clinicAver, degree: degreeAver, amputation: amputation}
		case uint(UR):
			urav = SearchData{bodyType: uint(UBODY), rom: romAver, clinic: clinicAver, degree: degreeAver, amputation: amputation}
		case uint(LL):
			llav = SearchData{bodyType: uint(LBODY), rom: romAver, clinic: clinicAver, degree: degreeAver, amputation: amputation}
		case uint(LR):
			lrav = SearchData{bodyType: uint(LBODY), rom: romAver, clinic: clinicAver, degree: degreeAver, amputation: amputation}
		case uint(TR):
			tr = SearchData{bodyType: uint(TBODY), rom: romAver}
		case uint(LOCOMOTION):
			loco = SearchData{bodyType: uint(LOCOBODY), rom: romAver}
		}

		fmt.Printf("BodyCompositionID: %d, ROM 평균: %d, Degree 평균: %d, Clinic 평균: %d\n", bodyCompId, romAver, degreeAver, clinicAver)
	}
	searchDatas := []SearchData{ulav, urav, llav, lrav}

	// var recommends []model.Recommended
	var recommends []model.Recommended
	query := service.db.Where("body_type_id = ? OR body_type_id = ? ", tr.bodyType, loco.bodyType)

	for _, v := range searchDatas {
		if v.amputation == 4 {
			query = query.Or("body_type_id = ? AND clinical_feature_id = ? AND degree_id <= ? AND amputation_id != 0 AND amputation_id <= 4 ", v.bodyType, v.clinic, v.degree)
		} else if v.amputation == 3 {
			query = query.Or("body_type_id = ? AND clinical_feature_id = ? AND degree_id <= ? AND amputation_id != 0 AND amputation_id <= 3", v.bodyType, v.clinic, v.degree)
		} else if v.amputation == 2 {
			query = query.Or("body_type_id = ? AND amputation_id != 0 AND amputation_id <= 2", v.bodyType)
		} else if v.amputation == 1 {
			query = query.Or("body_type_id = ? AND amputation_id = 1", v.bodyType)
		} else {
			query = query.Or("body_type_id = ? AND clinical_feature_id = ? AND degree_id <= ?", v.bodyType, v.clinic, v.degree)
		}
	}

	if err := query.Find(&recommends).Error; err != nil {
		return nil, errors.New("db error")
	}

	var recommendsTr, recommendsLoco, recommendsUl, recommendsUr, recommendsLl, recommendsLr []model.Recommended

	var originMap = make(map[uint][]model.Recommended)
	var recommendMap = make(map[uint][]model.Recommended)
	keys := []uint{uint(TR), uint(LOCOMOTION), uint(UL), uint(UR), uint(LL), uint(LR)}
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
			if ulav.amputation == 1 || ulav.amputation == 2 {
				originMap[uint(UL)] = append(originMap[uint(UL)], rec)
				recommendsUl = append(recommendsUl, rec)
				recommendMap[uint(UL)] = recommendsUl
			} else if urav.amputation == 1 || urav.amputation == 2 {
				originMap[uint(UR)] = append(originMap[uint(UR)], rec)
				recommendsUr = append(recommendsUr, rec)
				recommendMap[uint(UR)] = recommendsUr
			} else {
				if rec.ClinicalFeatureID == ulav.clinic && rec.DegreeID <= ulav.degree {
					if rec.IsAsymmetric && !(rec.ClinicalFeatureID == urav.clinic && rec.DegreeID <= urav.degree) {
						originMap[uint(UR)] = append(originMap[uint(UR)], rec)
						originMap[uint(UL)] = append(originMap[uint(UL)], rec)
					} else {
						originMap[uint(UL)] = append(originMap[uint(UL)], rec)
					}
				} else if rec.ClinicalFeatureID == urav.clinic && rec.DegreeID <= urav.degree {
					if rec.IsAsymmetric && !(rec.ClinicalFeatureID == ulav.clinic && rec.DegreeID <= ulav.degree) {
						originMap[uint(UL)] = append(originMap[uint(UL)], rec)
						originMap[uint(UR)] = append(originMap[uint(UR)], rec)
					} else {
						originMap[uint(UR)] = append(originMap[uint(UR)], rec)
					}
				}

				if rec.ClinicalFeatureID == ulav.clinic && rec.RomID <= ulav.rom && rec.DegreeID <= ulav.degree {
					recommendsUl = append(recommendsUl, rec)
					recommendMap[uint(UL)] = recommendsUl
				} else if rec.ClinicalFeatureID == urav.clinic && rec.RomID <= urav.rom && rec.DegreeID <= urav.degree {
					recommendsUr = append(recommendsUr, rec)
					recommendMap[uint(UR)] = recommendsUr
				}
			}

		case rec.BodyTypeID == llav.bodyType:
			if llav.amputation == 1 || llav.amputation == 2 {
				originMap[uint(LL)] = append(originMap[uint(LL)], rec)
				recommendsLl = append(recommendsLl, rec)
				recommendMap[uint(LL)] = recommendsLl
			} else if lrav.amputation == 1 || lrav.amputation == 2 {
				originMap[uint(LR)] = append(originMap[uint(LR)], rec)
				recommendsLr = append(recommendsLr, rec)
				recommendMap[uint(LR)] = recommendsLr
			} else {
				if rec.ClinicalFeatureID == llav.clinic && rec.DegreeID <= llav.degree {
					if rec.IsAsymmetric && !(rec.ClinicalFeatureID == lrav.clinic && rec.DegreeID <= lrav.degree) {
						originMap[uint(LR)] = append(originMap[uint(LR)], rec)
						originMap[uint(LL)] = append(originMap[uint(LL)], rec)
					} else {
						originMap[uint(LL)] = append(originMap[uint(LL)], rec)
					}
				} else if rec.ClinicalFeatureID == lrav.clinic && rec.DegreeID <= lrav.degree {
					if rec.IsAsymmetric && !(rec.ClinicalFeatureID == llav.clinic && rec.DegreeID <= llav.degree) {
						originMap[uint(LL)] = append(originMap[uint(LL)], rec)
						originMap[uint(LR)] = append(originMap[uint(LR)], rec)
					} else {
						originMap[uint(LR)] = append(originMap[uint(LR)], rec)
					}
				}

				if rec.ClinicalFeatureID == llav.clinic && rec.RomID <= llav.rom && rec.DegreeID <= llav.degree {
					recommendsLl = append(recommendsLl, rec)
					recommendMap[uint(LL)] = recommendsLl
				} else if rec.ClinicalFeatureID == lrav.clinic && rec.RomID <= lrav.rom && rec.DegreeID <= lrav.degree {
					recommendsLr = append(recommendsLr, rec)
					recommendMap[uint(LR)] = recommendsLr
				}
			}
		}
	}

	for _, key := range keys {
		if len(recommendMap[key]) == 0 {
			if len(originMap[key]) == 0 {
				log.Println("최종적으로 할수있는 운동 없음.")
				return nil, nil
			}
			log.Println("적합한 운동이 없어서 ", key, "번 bodycomposition rom 조건 포함안함.")
			recommendMap[key] = originMap[key]
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
