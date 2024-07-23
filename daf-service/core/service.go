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
	type SearchData struct {
		locomotion      uint
		bodyComposition uint
		jointAction     uint
		rom             uint
		clinic          uint
		degree          uint
		isGrip          bool
	}

	var searchDatas []SearchData
	var trom, locoRom uint

	for _, userAfc := range userAfcs {
		if userAfc.BodyCompositionID == uint(TR) {
			trom = *userAfc.RomID
		} else if userAfc.BodyCompositionID == uint(LOCOMOTION) {
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
					clinic: *userAfc.ClinicalFeatureID, degree: *degree, locomotion: locoRom})
			}
		}
	}

	var recommends []model.Recommended

	if err := service.db.Where("t_rom_id <= ? OR loco_rom_id <= ?", trom, locoRom).Find(&recommends).Error; err != nil {
		return nil, errors.New("db error")
	}

	if len(recommends) == 0 {
		log.Println("추천운동 없음")
		return nil, errors.New("-1")
	}

	// Extract recommend IDs
	recommendIDs := make([]uint, len(recommends))
	for i, recommend := range recommends {
		recommendIDs[i] = recommend.ID
	}

	// Step 2: Get RecommendedJointRom records based on recommended IDs
	var jointRoms []model.RecommendedJointRom
	if err := service.db.Where("recommended_id IN ?", recommendIDs).Find(&jointRoms).Error; err != nil {
		return nil, errors.New("db error2")
	}

	// Step 3: Get RecommendedClinicalDegree records based on recommended IDs
	var clinicalDegrees []model.RecommendedClinicalDegree
	if err := service.db.Where("recommended_id IN ?", recommendIDs).Find(&clinicalDegrees).Error; err != nil {
		return nil, errors.New("db error3")
	}

	recommendMap := make(map[uint][]uint)
	var asymmetrics []uint
	for _, afc := range searchDatas {
		for _, jointRom := range jointRoms {
			if afc.jointAction == uint(HIP) || afc.jointAction == uint(KNEE) {
				if afc.locomotion < 3 {
					if jointRom.JointActionID == uint(SUBHIP) || jointRom.JointActionID == uint(SUBKNEE) {
						if afc.rom >= jointRom.RomID {
							for _, clinicDegree := range clinicalDegrees {
								if afc.clinic == uint(MC) { // 힘은 추천에서 제외하기에
									for _, recommend := range recommends {
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
					continue
				}
			}
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

	recommendIds := mergeAndRemoveDuplicates(interserctionAll, asymmetrics)

	var exerciseIds []uint
	for _, v := range recommends {
		for _, w := range recommendIds {
			if v.ID == w {
				exerciseIds = append(exerciseIds, v.ExerciseID)
			}
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
	if len(recommendIds) == 0 {
		log.Println("교집합이 없어서 최종적으로 할수있는 운동 없음")
		return nil, nil
	} else if len(recommendIds) <= RECOMMENDCOUNT {
		log.Println("운동 추천완료")
		var exercises []model.Exercise
		if err := service.db.Where("id IN ? ", exerciseIds).Find(&exercises).Error; err != nil {
			return nil, errors.New("db error2")
		}

		for _, id := range categoryIds {
			for _, exercise := range exercises {
				if exercise.CategoryId == id {
					log.Println("데일리 카테고리에 해당하는지 분류")
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
		if err := service.db.Where("id IN ? ", exerciseIds).Find(&exercises).Error; err != nil {
			return nil, errors.New("db error2")
		}
		if err := service.db.Where("exercise_id IN ? ", exerciseIds).Find(&histories).Error; err != nil {
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
