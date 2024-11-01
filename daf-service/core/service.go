// /user-service/service/service.go

package core

import (
	"daf-service/model"
	"errors"
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

	var searchDatas []SearchData
	var locoRom uint

	for _, userAfc := range userAfcs {
		if userAfc.BodyCompositionID == uint(LOCOMOTION) {
			locoRom = *userAfc.RomID
		} else {
			if userAfc.JointActionID != uint(WRIST) && userAfc.JointActionID != uint(FINGER) && userAfc.JointActionID != uint(ANKLE) {
				rom := userAfc.RomID
				degree := userAfc.DegreeID
				isGrip := userAfc.IsGrip
				if rom == nil {
					temp := uint(1)
					rom = &temp
				}
				if degree == nil {
					temp := uint(1)
					degree = &temp
				}
				if isGrip == nil {
					temp := true
					isGrip = &temp
				}
				searchDatas = append(searchDatas, SearchData{bodyComposition: userAfc.BodyCompositionID, jointAction: userAfc.JointActionID, rom: *rom, isGrip: *isGrip,
					clinic: *userAfc.ClinicalFeatureID, degree: *degree})
			}
		}
	}

	var recommends []model.Recommended

	if err := service.db.Where("loco_rom_id <= ?", locoRom).Find(&recommends).Error; err != nil {
		return nil, errors.New("db error")
	}

	if len(recommends) == 0 {
		log.Println("추천운동 없음")
		return nil, errors.New("-1")
	}

	//recommendIDs 추출
	recommendIDs := make([]uint, len(recommends))
	for i, recommend := range recommends {
		recommendIDs[i] = recommend.ID
	}

	// RecommendedJointRom 레코드 based on recommendedIDs
	var jointRoms []model.RecommendedJointRom
	if err := service.db.Where("recommended_id IN ?", recommendIDs).Find(&jointRoms).Error; err != nil {
		return nil, errors.New("db error2")
	}

	// RecommendedClinicalDegree 레코드 based on recommendedIDs
	var clinicalDegrees []model.RecommendedClinicalDegree
	if err := service.db.Where("recommended_id IN ?", recommendIDs).Find(&clinicalDegrees).Error; err != nil {
		return nil, errors.New("db error3")
	}

	recommendMap := make(map[uint][]uint)
	var asymmetrics []uint
	for _, afc := range searchDatas {
		for _, jointRom := range jointRoms {
			if afc.jointAction == jointRom.JointActionID {
				if afc.rom >= jointRom.RomID {
					for _, clinicDegree := range clinicalDegrees {
						if afc.clinic == uint(MC) { // 힘은 추천에서 제외하기에
							for _, recommend := range recommends {
								if !afc.isGrip && recommend.IsGrip != nil {
									if *recommend.IsGrip {
										continue
									}
								}
								if clinicDegree.RecommendedID == recommend.ID {
									if recommend.IsAsymmetric {
										asymmetrics = append(asymmetrics, recommend.ID)
									} else {
										recommendMap[afc.bodyComposition] = append(recommendMap[afc.bodyComposition], clinicDegree.RecommendedID)
									}
								}
							}
						}
						if afc.clinic == clinicDegree.ClinicalFeatureID {
							if afc.degree >= clinicDegree.DegreeID {
								for _, recommend := range recommends {
									if !afc.isGrip && recommend.IsGrip != nil {
										if *recommend.IsGrip {
											continue
										}
									}
									if clinicDegree.RecommendedID == recommend.ID {
										if recommend.IsAsymmetric {
											asymmetrics = append(asymmetrics, recommend.ID)
										} else {
											recommendMap[afc.bodyComposition] = append(recommendMap[afc.bodyComposition], clinicDegree.RecommendedID)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	//오,왼의 교집합 -> 위,아래 교집합 + 각 에스메트릭 운동
	intersectionU := intersect(recommendMap[uint(UL)], recommendMap[uint(UR)])
	intersectionL := intersect(recommendMap[uint(LL)], recommendMap[uint(LR)])
	interserctionAll := intersect(intersectionU, intersectionL)
	interserctionALLTr := intersect(interserctionAll, recommendMap[uint(TR)])

	recommendIds := mergeAndRemoveDuplicates(interserctionALLTr, asymmetrics)

	var exerciseIds []uint
	for _, v := range recommends {
		for _, w := range recommendIds {
			if v.ID == w {
				exerciseIds = append(exerciseIds, v.ExerciseID)
			}
		}
	}

	//daily 카테고리 선별 정책 필요함 정책 확정 전까지 임시로 모든 카테고리
	var categoris []model.Category
	if err := service.db.Find(&categoris).Error; err != nil {
		return nil, errors.New("db error4")
	}

	categoryIds := []uint{}
	for _, c := range categoris {
		categoryIds = append(categoryIds, c.ID)
	}

	result := make(map[uint]RecomendResponse)
	if len(recommendIds) == 0 {
		log.Println("교집합이 없어서 최종적으로 할수있는 운동 없음")
		return nil, nil
	} else if len(recommendIds) <= RECOMMENDCOUNT {
		log.Println("운동 추천완료")
		var categoryExercises []model.CategoryExercise
		if err := service.db.Where("exercise_id IN ? ", exerciseIds).Preload("Exercise").Find(&categoryExercises).Error; err != nil {
			return nil, errors.New("db error2")
		}

		for _, id := range categoryIds {
			for _, categoryExercise := range categoryExercises {
				if categoryExercise.CategoryID == id {
					log.Println("데일리 카테고리에 해당하는지 분류")
					ex := ExerciseResponse{ID: categoryExercise.ExerciseID, Name: categoryExercise.Exercise.Name}
					r := result[uint(id)]
					r.First = append(r.First, ex)
					result[uint(id)] = r
				}
			}
		}

	} else {
		log.Println("교집합이 많음")
		var categoryExercises []model.CategoryExercise
		var uniqueExerciseIds []uint
		if err := service.db.Where("exercise_id IN ? ", exerciseIds).Preload("Exercise").Find(&categoryExercises).Error; err != nil {
			return nil, errors.New("db error2")
		}
		if err := service.db.Model(&model.History{}).
			Select("DISTINCT exercise_id").
			Where("exercise_id IN ?", exerciseIds).
			Pluck("exercise_id", &uniqueExerciseIds).Error; err != nil {
			return nil, errors.New("db error3")
		}

		exerciseCount := make(map[uint]int)
		for _, id := range uniqueExerciseIds {
			exerciseCount[id]++
		}

		for _, id := range categoryIds {
			for _, categoryExercise := range categoryExercises {
				if categoryExercise.CategoryID == id {
					log.Println("데일리 카테고리에 해당하는지 분류")
					ex := ExerciseResponse{ID: categoryExercise.Exercise.ID, Name: categoryExercise.Exercise.Name}
					r := result[id]
					r.First = append(r.First, ex)
					result[id] = r
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

type SearchData struct {
	bodyComposition uint
	jointAction     uint
	rom             uint
	clinic          uint
	degree          uint
	isGrip          bool
}