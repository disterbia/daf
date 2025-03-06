package core

import (
	"admin-service/model"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

var jwtSecretKey = []byte("haruharu_mark_admin")

func verifyJWT(c *gin.Context) (uint, string, error) {
	// 헤더에서 JWT 토큰 추출
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return 0, "", errors.New("authorization header is required")
	}

	// 'Bearer ' 접두사 제거
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if err != nil || !token.Valid {
		return 0, "", errors.New("invalid token")
	}

	id := uint((*claims)["id"].(float64))
	email := (*claims)["email"].(string)
	if email == "" || id == 0 {
		return 0, "", errors.New("id or email not found in token")
	}
	return id, email, nil
}

func generateJWT(admin model.Admin) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    admin.ID,
		"email": admin.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(), // 한달 유효 기간
	})

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func copyStruct(src, dst interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldName := srcVal.Type().Field(i).Name

		dstField := dstVal.FieldByName(srcFieldName)
		if dstField.IsValid() && dstField.Type() == srcField.Type() {
			dstField.Set(srcField)
		}
	}

	return nil
}

func validateEmail(email string) error {
	// 빈 문자열 검사
	if email == "" || len(email) > 50 || strings.Contains(email, " ") {
		return errors.New("invalid email format")
	}

	// 이메일 검증을 위한 정규 표현식
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

func validateSignIn(request SignInRequest) error {
	pattern := `^010\d{8}$`
	matched, err := regexp.MatchString(pattern, request.Phone)
	if err != nil || !matched {
		return errors.New("invalid phone format, should be 01000000000")
	}

	name := strings.TrimSpace(request.Name)
	if utf8.RuneCountInString(name) > 50 || len(name) == 0 {
		return errors.New("invalid name")
	}
	return nil
}

func validateSaveUser(request SaveUserRequest) error {
	pattern := `^010\d{8}$`
	matched, err := regexp.MatchString(pattern, request.Phone)
	if err != nil || !matched {
		return errors.New("invalid phone format, should be 01000000000")
	}

	name := strings.TrimSpace(request.Name)
	if utf8.RuneCountInString(name) > 50 || len(name) == 0 {
		return errors.New("invalid name")
	}
	return nil
}

func contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func checkDuplicates(slice []uint) bool {
	seen := make(map[uint]struct{})
	for _, item := range slice {
		if _, exists := seen[item]; exists {
			return true // Duplicate found
		}
		seen[item] = struct{}{}
	}
	return false // No duplicates
}

func calculateAgeCode(birthday time.Time) uint {
	now := time.Now()
	age := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		age--
	}

	switch {
	case age < 20:
		return 1
	case age < 30:
		return 2
	case age < 40:
		return 3
	case age < 50:
		return 4
	case age < 60:
		return 5
	case age < 70:
		return 6
	default:
		return 7
	}
}

func getBirthdayRangeByAgeCode(ageCode uint) (time.Time, time.Time, error) {
	now := time.Now()
	currentYear := now.Year()
	var startYear, endYear int

	switch ageCode {
	case 1: // 0-9 years
		startYear = currentYear - 9
		endYear = currentYear
	case 2: // 10-19 years
		startYear = currentYear - 19
		endYear = currentYear - 10
	case 3: // 20-29 years
		startYear = currentYear - 29
		endYear = currentYear - 20
	case 4: // 30-39 years
		startYear = currentYear - 39
		endYear = currentYear - 30
	case 5: // 40-49 years
		startYear = currentYear - 49
		endYear = currentYear - 40
	case 6: // 50-59 years
		startYear = currentYear - 59
		endYear = currentYear - 50
	case 7: // 60 years and above
		startYear = 0
		endYear = currentYear - 60
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid age code")
	}

	startDate := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(endYear, 12, 31, 23, 59, 59, 999999999, time.UTC)
	return startDate, endDate, nil
}

func validateAfc(request []AfcRequest) string {
	// if len(request) != 16 {
	// 	return "all parts must fill"
	// }

	checkBodyJoint := make(map[uint]map[uint]bool)
	checkIsGrip := make(map[uint]*bool)

	for _, v := range request {
		// BodyCompositionID 0을 허용하지 않음
		if v.BodyCompositionID == 0 {
			return "BodyCompositionID cannot be 0"
		}

		if checkBodyJoint[v.BodyCompositionID] == nil {
			checkBodyJoint[v.BodyCompositionID] = make(map[uint]bool)
		}

		if checkBodyJoint[v.BodyCompositionID][v.JointActionID] {
			return "duplicate parts"
		}

		if checkIsGrip[v.BodyCompositionID] != nil {
			if checkIsGrip[v.BodyCompositionID] != &v.IsGrip {
				return "is_grip must same"
			}
		}

		if v.BodyCompositionID != uint(LOCOMOTION) {
			if v.Pain == 0 || v.Pain > 5 {
				return "check pain"
			}
		}

		// BodyCompositionID와 JointActionID 짝 검사
		validJointAction := false
		switch v.BodyCompositionID {
		case uint(LOCOMOTION):
			// LOCOMOTION일 때 JointActionID는 0이어야 함
			validJointAction = v.JointActionID == 0
		case uint(TR):
			validJointAction = v.JointActionID == uint(TR)
		case uint(UL), uint(UR):
			if v.JointActionID != uint(FINGER) && v.ClinicalFeatureID == uint(AC) && v.IsGrip {
				return "isGrip must false"
			}
			validJointAction = v.JointActionID == uint(SHOULDER) || v.JointActionID == uint(ELBOW) || v.JointActionID == uint(WRIST) || v.JointActionID == uint(FINGER)
		case uint(LL), uint(LR):
			validJointAction = v.JointActionID == uint(HIP) || v.JointActionID == uint(KNEE) || v.JointActionID == uint(ANKLE)
		default:
			return "Invalid BodyCompositionID"
		}

		if !validJointAction {
			return "Invalid JointActionID for given BodyCompositionID"
		}

		if v.BodyCompositionID == uint(LOCOMOTION) {
			// LOCOMOTION일 때
			if v.RomID == 0 {
				return "RomID must not be 0 when BodyCompositionID is LOCOMOTION"
			}
			if v.ClinicalFeatureID != 0 || v.DegreeID != 0 {
				return "ClinicalFeatureID and DegreeID must be 0 when BodyCompositionID is  LOCOMOTION"
			}
		} else {
			//  LOCOMOTION이 아닐 때
			if v.JointActionID == 0 {
				return "JointActionID cannot be 0 when BodyCompositionID is not  LOCOMOTION"
			}
			if v.ClinicalFeatureID == uint(AC) {
				// ClinicalFeatureID가 AC일 때
				if v.RomID != 0 || v.DegreeID != 0 {
					return "RomID and DegreeID must be 0 when ClinicalFeatureID is AC"
				}
			} else {
				// ClinicalFeatureID가 AC가 아닐 때
				if v.RomID == 0 || v.ClinicalFeatureID == 0 || v.DegreeID == 0 {
					return "RomID, ClinicalFeatureID, and DegreeID cannot be 0 when JointActionID is 4 or less and ClinicalFeatureID is not AC"
				}
			}
			if v.ClinicalFeatureID == uint(MC) {
				if !(v.DegreeID == 1 || v.DegreeID == 5) {
					return "muscular force must 1 or 5"
				}
			}

		}
		checkBodyJoint[v.BodyCompositionID][v.JointActionID] = true
		checkIsGrip[v.BodyCompositionID] = &v.IsGrip
	}

	return "" // 모든 검증을 통과하면 빈 문자열 반환
}
func sum(slice []uint) uint {
	total := uint(0)
	for _, value := range slice {
		total += value
	}
	return total
}

func checkFamily(service *adminService, adminId, userId uint) (bool, error) {
	var superAgencyID uint
	if err := service.db.Table("admins").
		Select("agencies.super_agency_id").
		Joins("JOIN agencies ON agencies.id = admins.agency_id").
		Where("admins.id = ?", adminId).
		Scan(&superAgencyID).Error; err != nil {
		return false, errors.New("db error: could not find admin's super agency")
	}

	// Step 2: Check if the user's agency is under the admin's superAgency
	var count int64
	if err := service.db.Table("users").
		Joins("JOIN agencies AS user_agency ON user_agency.id = users.agency_id").
		Joins("JOIN agencies AS admin_agency ON admin_agency.super_agency_id = ?", superAgencyID).
		Where("users.id = ? AND user_agency.super_agency_id = admin_agency.super_agency_id", userId).
		Count(&count).Error; err != nil {
		return false, errors.New("db error: could not check user's agency")
	}

	return count > 0, nil
}

func validateDiary(request SaveDiaryRequest) bool {
	if request.Uid == 0 || request.Title == "" || request.ClassDate == "" || request.ClassName == "" || len(request.ClassPurposeIDs) == 0 || request.ClassType == 0 ||
		len(request.ExerciseMeasures) == 0 {
		return false
	}
	checkDuplicate := make(map[uint]bool)
	for _, v := range request.ExerciseMeasures {
		for _, w := range v.Measures {
			if checkDuplicate[w.MeasureID] {
				return false
			}
			checkDuplicate[w.MeasureID] = true
		}
	}
	return true
}

func deleteFromS3(fileKey string, s3Client *s3.S3, bucket string, bucketUrl string) error {

	// URL에서 객체 키 추출
	key := extractKeyFromUrl(fileKey, bucket, bucketUrl)
	log.Println("key", fileKey)

	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	// 에러 발생 시 처리 로직
	if err != nil {
		fmt.Printf("Failed to delete object from S3: %s, error: %v\n", fileKey, err)
	}

	return err
}

// URL에서 S3 객체 키를 추출하는 함수
func extractKeyFromUrl(url, bucket string, bucketUrl string) string {
	prefix := fmt.Sprintf("https://%s.%s/", bucket, bucketUrl)
	return strings.TrimPrefix(url, prefix)
}

func uploadImagesToS3(imgData []byte, contentType string, ext string, s3Client *s3.S3, bucket string, bucketUrl string, uid string) (string, error) {
	// 이미지 파일 이름과 썸네일 파일 이름 생성
	imgFileName := "images/recommend/" + uid + "/images/" + uuid.New().String() + ext

	// S3에 이미지 업로드
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(imgFileName),
		Body:        bytes.NewReader(imgData),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", err
	}

	// 업로드된 이미지와 썸네일의 URL 생성 및 반환
	imgURL := "https://" + bucket + "." + bucketUrl + "/" + imgFileName

	return imgURL, nil
}

func reduceImageSize(data []byte) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	log.Println("image size: ", len(data))
	// 원본 이미지의 크기를 절반씩 줄이면서 10MB 이하로 만듦
	for len(data) > 10*1024*1024 {
		newWidth := img.Bounds().Dx() / 2
		newHeight := img.Bounds().Dy() / 2

		resizedImg := resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)

		var buf bytes.Buffer
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, resizedImg, nil)
		case "png":
			err = png.Encode(&buf, resizedImg)
		case "gif":
			err = gif.Encode(&buf, resizedImg, nil)
		case "webp":
			// WebP 인코딩은 지원하지 않으므로 PNG 형식으로 인코딩
			err = png.Encode(&buf, resizedImg)
		// 여기에 필요한 다른 형식을 추가할 수 있습니다.
		default:
			log.Printf("Unsupported format: %s\n", format)
			return nil, err
		}
		if err != nil {
			return nil, err
		}

		data = buf.Bytes()
		img = resizedImg
	}

	return data, nil
}

func getImageFormat(imgData []byte) (contentType, extension string, err error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(imgData))
	if err != nil {
		return "", "", err
	}
	switch format {
	case "jpeg":
		contentType = "image/jpeg"
		extension = ".jpg"
	case "png":
		contentType = "image/png"
		extension = ".png"
	case "gif":
		contentType = "image/gif"
		extension = ".gif"
	case "wepb":
		contentType = "image/wepb"
		extension = ".wepb"
	default:
		return "", "", fmt.Errorf("unsupported image format: %s", format)
	}

	return contentType, extension, nil
}
