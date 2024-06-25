package model

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 데이터베이스 연결 초기화
func NewDB(dataSourceName string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err = sqlDB.Ping(); err != nil {
		return nil, err
	}

	db.AutoMigrate(&Admin{}, &Agency{}, &AppVersion{}, &AuthCode{}, &BodyComposition{}, &BodyType{}, &Category{}, &ClinicalFeature{}, &Degree{}, &DisableDetail{}, &DisableType{}, &ExerciseMachine{},
		&ExerciseMeasure{}, &ExercisePurpose{}, &Exercise{}, &History{}, &Image{}, &JointAction{}, &Machine{}, &Measure{}, &Purpose{}, &Recommended{}, &Role{}, &Rom{}, &UseStatus{},
		&UserDisableDetail{}, &UserDisable{}, &UserAfc{}, &UserVisit{}, &User{}, &VerifiedEmail{}, &VisitPurpose{})
	return db, nil
}
