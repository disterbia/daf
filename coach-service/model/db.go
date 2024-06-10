package model

import (
	"time"

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

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	db.AutoMigrate(&BodyComposition{}, &Category{}, &ClinicalFeature{}, &Degree{}, &Exercise{}, &History{}, &JointAction{}, &Recommended{}, &Rom{}, &UserJointAction{},
		&Machine{}, &ExerciseMachine{}, &Purpose{}, &ExercisePurpose{}, &BodyType{}, &Admin{}, &Agency{})
	// createHangulFunctions(db)
	return db, nil

}

// func createHangulFunctions(db *gorm.DB) {
// 	functionSQL := `
// 	CREATE OR REPLACE FUNCTION get_chosung(input_str TEXT)
// 	RETURNS TEXT AS $$
// 	DECLARE
// 		result TEXT = '';
// 		char_code INT;
// 		chosung CHAR(1);
// 	BEGIN
// 		FOR i IN 1..char_length(input_str) LOOP
// 			char_code := get_byte(convert_to(substr(input_str, i, 1), 'UTF8'), 0);
// 			IF char_code BETWEEN 0 AND 127 THEN
// 				-- ASCII 문자
// 				result := result || substr(input_str, i, 1);
// 			ELSE
// 				-- 한글 문자
// 				char_code := ((char_code - 224) * 4096) + ((get_byte(convert_to(substr(input_str, i, 1), 'UTF8'), 1) - 128) * 64) + (get_byte(convert_to(substr(input_str, i, 1), 'UTF8'), 2) - 128);
// 				IF char_code BETWEEN 44032 AND 55203 THEN
// 					chosung := chr(((char_code - 44032) / 588) + 4352);
// 					result := result || chosung;
// 				ELSE
// 					result := result || substr(input_str, i, 1);
// 				END IF;
// 			END IF;
// 		END LOOP;
// 		RETURN result;
// 	END;
// 	$$ LANGUAGE plpgsql;
//     `

// 	if err := db.Exec(functionSQL).Error; err != nil {
// 		log.Fatalf("Failed to create function: %v", err)
// 	} else {
// 		log.Println("Function get_chosung created successfully")
// 	}
// }
