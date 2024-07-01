// /user-service/service/service.go

package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"admin-service/model"
	pb "admin-service/proto"

	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type AdminService interface {
	sendAuthCode(number string) (string, error)
	login(request LoginRequest) (string, error)
	verifyAuthCode(verify VerifyRequest) (string, error)
	signIn(request SignInRequest) (string, error)
	resetPassword(request LoginRequest) (string, error)
	saveUser(request SaveUserRequest) (string, error)
	searchUsers(request SearchUserRequest) ([]SearchUserResponse, error)
	getAgencis(id uint) ([]AgAdResponse, error)
	getAdmins(id uint) ([]AgAdResponse, error)
	getDisableDetails() ([]AgAdResponse, error)
	getAfcs(id, uid uint) (GetAfcResponse, error)
	createAfc(request SaveAfcRequest) (string, error)
	updateAfc(request SaveAfcRequest) (string, error)
	getAfcHistoris(id, uid uint) ([]GetAfcResponse, error)
	updateAfcHistory(request SaveAfcHistoryRequest) (string, error)

	searchDiary(request SearchDiaryRequest) ([]SearchDiaryResponse, error)
	saveDiary(request SaveDiaryRequest) (string, error)
}

type adminService struct {
	db          *gorm.DB
	emailClient pb.EmailServiceClient
	s3svc       *s3.S3
	bucket      string
	bucketUrl   string
}

func NewAdminService(db *gorm.DB, conn *grpc.ClientConn, s3svc *s3.S3, bucket string, bucketUrl string) AdminService {
	emailClient := pb.NewEmailServiceClient(conn)
	return &adminService{db: db, emailClient: emailClient, s3svc: s3svc, bucket: bucket, bucketUrl: bucketUrl}
}

func (service *adminService) login(request LoginRequest) (string, error) {
	var u model.Admin
	password := strings.TrimSpace(request.Password)

	if password == "" {
		return "", errors.New("empty")
	}

	// 이메일로 사용자 조회
	if err := service.db.Where("email = ?", request.Email).First(&u).Error; err != nil {
		return "", err
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

func (service *adminService) sendAuthCode(email string) (string, error) {
	log.Println(email)
	err := validateEmail(email)
	if err != nil {
		return "", err
	}

	// result := service.db.Where("email=?", email).Find(&model.Admin{})
	// if result.Error != nil {
	// 	return "", errors.New("db error")

	// } else if result.RowsAffected > 0 {
	// 	// 레코드가 존재할 때
	// 	return "", errors.New("-1")
	// }

	var sb strings.Builder
	for i := 0; i < 6; i++ {
		fmt.Fprintf(&sb, "%d", rand.Intn(10)) // 0부터 9까지의 숫자를 무작위로 선택
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic in goroutine:", r)
			}
		}()
		response, err := service.emailClient.SendEmail(context.Background(), &pb.EmailRequest{
			Email: email, // 받는 사람의 이메일
			Code:  sb.String(),
		})
		if err != nil {
			log.Println(err)
		}
		if response != nil && response.Status == "Success" {
			if err := service.db.Create(&model.AuthCode{Email: email, Code: sb.String()}).Error; err != nil {
				log.Println(err)
			}
		}
	}()

	return "200", nil
}

func (service *adminService) verifyAuthCode(verify VerifyRequest) (string, error) {
	var authCode model.AuthCode

	if err := service.db.Where("email = ? ", verify.Email).Last(&authCode).Error; err != nil {
		return "", errors.New("db error")
	}
	if authCode.Code != verify.Code {
		return "", errors.New("-1")
	}

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	if err := tx.Where("email = ?", authCode.Email).Unscoped().Delete(&authCode).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error3")
	}

	if err := tx.Create(&model.VerifiedEmail{Email: authCode.Email}).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}

	tx.Commit()
	return "200", nil
}

func (service *adminService) signIn(request SignInRequest) (string, error) {
	if err := validateSignIn(request); err != nil {
		return "", err
	}
	// 비밀번호 공백 제거
	password := strings.TrimSpace(request.Password)

	if password == "" {
		return "", errors.New("empty")
	}

	var verify model.VerifiedEmail
	result := service.db.Where("email=?", request.Email).Find(&verify)
	if result.Error != nil {
		return "", errors.New("db error")

	} else if result.RowsAffected == 0 {
		// 인증안함
		return "", errors.New("-1")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	var admin = model.Admin{}

	if err := copyStruct(request, &admin); err != nil {
		return "", err
	}

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	if err := tx.Where("email = ?", request.Email).Unscoped().Delete(&verify).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}

	admin.Password = string(hashedPassword)
	admin.RoleID = 1
	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error3")
	}
	tx.Commit()
	return "200", nil
}

func (service *adminService) resetPassword(request LoginRequest) (string, error) {
	// 비밀번호 공백 제거
	password := strings.TrimSpace(request.Password)

	if password == "" {
		return "", errors.New("empty")
	}

	var verify model.VerifiedEmail
	result := service.db.Where("email=?", request.Email).Find(&verify)
	if result.Error != nil {
		return "", errors.New("db error")

	} else if result.RowsAffected == 0 {
		// 인증안함
		return "", errors.New("-1")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	var admin = model.Admin{}

	if err := copyStruct(request, &admin); err != nil {
		return "", err
	}

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	if err := tx.Where("email = ?", request.Email).Unscoped().Delete(&verify).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}

	if err := tx.Model(&admin).Where("email = ?", request.Email).Update("password", string(hashedPassword)).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error3")
	}
	tx.Commit()
	return "200", nil
}

func (service *adminService) saveUser(request SaveUserRequest) (string, error) {
	if result, err := checkFamily(service, request.Uid, request.ID); err != nil || !result {
		if !result {
			return "", errors.New("not family")
		}
	}

	if err := validateSaveUser(request); err != nil {
		return "", err
	}
	if len(request.VisitPurposeIDs) == 0 {
		return "", errors.New("visit err")
	}
	if len(request.DisableTypeIDs) == 0 {
		return "", errors.New("disable err")
	}
	if checkDuplicates(request.DisableTypeIDs) {
		return "", errors.New("duplicate Id1")
	}
	if checkDuplicates(request.VisitPurposeIDs) {
		return "", errors.New("duplicate Id2")
	}
	if checkDuplicates(request.DisableDetailIDs) {
		return "", errors.New("duplicate Id3")
	}
	birthday, err := time.Parse("2006-01-02", request.Birthday)
	if err != nil {
		return "", errors.New("date err1")
	}
	registday, err := time.Parse("2006-01-02", request.RegistDay)
	if err != nil {
		return "", errors.New("date err2")
	}

	var vists []model.UserVisit
	var disables []model.UserDisable
	var details []model.UserDisableDetail
	var user model.User

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	// 기존 사용자 로드 또는 새 사용자 생성
	if err := service.db.Where("id = ?", request.ID).First(&user).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("db error")
		}
	}

	if err := copyStruct(request, &user); err != nil {
		return "", err
	}

	user.CreateAdminID = request.Uid
	user.Birthday = birthday
	user.RegistDay = registday

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error1")
	}

	//방문목적 저장
	for _, v := range request.VisitPurposeIDs {
		vists = append(vists, model.UserVisit{UID: user.ID, VisitPurposeID: v})
	}
	if err := tx.Unscoped().Where("uid = ?", user.ID).Delete(&vists).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}
	if err := tx.Create(&vists).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error3")
	}

	//장애유형 저장
	for _, v := range request.DisableTypeIDs {
		disables = append(disables, model.UserDisable{UID: user.ID, DisableTypeID: v})
	}
	if err := tx.Unscoped().Where("uid = ?", user.ID).Delete(&disables).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error4")
	}
	if err := tx.Create(&disables).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error5")
	}

	//장애유형 기타 저장
	if contains(request.DisableTypeIDs, uint(ETC)) && len(request.DisableTypeIDs) != 0 {
		for _, v := range request.DisableDetailIDs {
			details = append(details, model.UserDisableDetail{UID: user.ID, DisableDetailID: v})
		}
		if err := tx.Unscoped().Where("uid = ?", user.ID).Delete(&details).Error; err != nil {
			tx.Rollback()
			return "", errors.New("db error6")
		}
		if err := tx.Create(&details).Error; err != nil {
			tx.Rollback()
			return "", errors.New("db error7")
		}
	}
	tx.Commit()
	return "200", nil
}

func (service *adminService) searchUsers(request SearchUserRequest) ([]SearchUserResponse, error) {

	var superAgencyID uint
	if err := service.db.Table("admins").
		Select("agencies.super_agency_id").
		Joins("JOIN agencies ON agencies.id = admins.agency_id").
		Where("admins.id = ?", request.Id).
		Scan(&superAgencyID).Error; err != nil {
		return nil, errors.New("db error")
	}

	pageSize := 10
	offset := int(request.Page) * pageSize

	var users []model.User
	query := service.db.Model(&model.User{}).
		Joins("JOIN agencies ON agencies.id = users.agency_id").
		Where("agencies.super_agency_id = ?", superAgencyID)

	if strings.TrimSpace(request.Name) != "" {
		query = query.Where("users.name LIKE ?", "%"+request.Name+"%")
	}
	if request.Gender != 0 {
		query = query.Where("users.gender = ?", request.Gender)
	}
	if request.AgencyID != 0 {
		query = query.Where("users.agency_id = ?", request.AgencyID)
	}
	if request.AdminID != 0 {
		query = query.Where("users.admin_id = ?", request.AdminID)
	}
	if request.UseStatusID != 0 {
		query = query.Where("users.use_status_id = ?", request.UseStatusID)
	}
	if len(request.DisableTypeIDs) > 0 {
		query = query.Where("users.id IN (SELECT uid FROM user_disables WHERE disable_type_id IN ?)", request.DisableTypeIDs)
	}
	if len(request.VisitPurposeIDs) > 0 {
		query = query.Where("users.id IN (SELECT uid FROM user_visits WHERE visit_purpose_id IN ?)", request.VisitPurposeIDs)
	}
	if len(request.DisableDetailIDs) > 0 {
		query = query.Where("users.id IN (SELECT uid FROM user_disable_details WHERE disable_detail_id IN ?)", request.DisableDetailIDs)
	}
	// Afc 조건 추가
	if len(request.Afcs) > 0 {
		for _, afc := range request.Afcs {
			if afc.BodyCompositionID != 0 {
				query = query.Where("users.id IN (SELECT uid FROM user_joint_actions WHERE body_composition_id = ?)", afc.BodyCompositionID)
			}
			if afc.JointActionID != 0 {
				query = query.Where("users.id IN (SELECT uid FROM user_joint_actions WHERE joint_action_id = ?)", afc.JointActionID)
			}
			if afc.RomID != 0 {
				query = query.Where("users.id IN (SELECT uid FROM user_joint_actions WHERE rom_id = ?)", afc.RomID)
			}
			if afc.ClinicalFeatureID != 0 {
				query = query.Where("users.id IN (SELECT uid FROM user_joint_actions WHERE clinical_feature_id = ?)", afc.ClinicalFeatureID)
			}
			if afc.DegreeID != 0 {
				query = query.Where("users.id IN (SELECT uid FROM user_joint_actions WHERE degree_id = ?)", afc.DegreeID)
			}
		}
	}

	if request.AgeCode != 0 {
		startDate, endDate, err := getBirthdayRangeByAgeCode(request.AgeCode)
		if err != nil {
			return nil, err
		}
		query = query.Where("users.birthday BETWEEN ? AND ?", startDate, endDate)
	}

	if strings.TrimSpace(request.RegistDay) != "" {
		registDay, err := time.Parse("2006-01-02", request.RegistDay)
		if err != nil {
			return nil, errors.New("invalid date format for registDay")
		}
		query = query.Where("users.regist_day = ?", registDay)
	}

	query = query.Offset(offset).Limit(pageSize)

	if err := query.Preload("Agency").Preload("Admin").Preload("UseStatus").Find(&users).Error; err != nil {
		return nil, errors.New("db error")
	}

	// 사용자 ID 리스트 가져오기
	userIDs := make([]uint, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	// DisableTypeNames, VisitPurposeNames, DisableDetailNames을 가져오기 위한 쿼리
	var userDisables []model.UserDisable
	var userVisits []model.UserVisit
	var userDisableDetails []model.UserDisableDetail
	var userAfcs []model.UserAfc

	service.db.Where("uid IN ?", userIDs).Preload("DisableType").Find(&userDisables)
	service.db.Where("uid IN ?", userIDs).Preload("VisitPurpose").Find(&userVisits)
	service.db.Where("uid IN ?", userIDs).Preload("DisableDetail").Find(&userDisableDetails)
	service.db.Where("uid IN ?", userIDs).Preload("ClinicalFeature").Find(&userAfcs)
	// 사용자 ID를 키로 하는 맵 생성
	disableTypeNamesMap := make(map[uint][]IdNameResponse)
	visitPurposeNamesMap := make(map[uint][]IdNameResponse)
	disableDetailNamesMap := make(map[uint][]IdNameResponse)
	userJointActionsMap := make(map[uint][]AfcResponse)
	type GroupData struct {
		romList    []uint
		clinicList []string
		degreeList []uint
	}
	groupData := make(map[uint]map[uint]GroupData)

	for _, ud := range userDisables {
		disableTypeNamesMap[ud.UID] = append(disableTypeNamesMap[ud.UID], IdNameResponse{Id: ud.DisableTypeID, Name: ud.DisableType.Name})
	}

	for _, uv := range userVisits {
		visitPurposeNamesMap[uv.UID] = append(visitPurposeNamesMap[uv.UID], IdNameResponse{Id: uv.VisitPurposeID, Name: uv.VisitPurpose.Name})
	}

	for _, udd := range userDisableDetails {
		disableDetailNamesMap[udd.UID] = append(disableDetailNamesMap[udd.UID], IdNameResponse{Id: udd.DisableDetailID, Name: udd.DisableDetail.Name})
	}

	for _, userAfc := range userAfcs {
		// BodyCompositionID를 키로 사용하는 그룹에 데이터 추가
		uid := userAfc.Uid
		if groupData[uid] == nil {
			groupData[uid] = make(map[uint]GroupData)
		}
		userMap := groupData[uid]
		bodyCompId := userAfc.BodyCompositionID

		data := userMap[bodyCompId]
		if userAfc.RomID != nil {
			data.romList = append(data.romList, *userAfc.RomID)
		}
		if userAfc.ClinicalFeatureID != nil {
			data.clinicList = append(data.clinicList, userAfc.ClinicalFeature.Code)
		}
		if userAfc.DegreeID != nil {
			data.degreeList = append(data.degreeList, *userAfc.DegreeID)
		}

		groupData[uid][bodyCompId] = data

	}

	// 그룹별 평균 계산
	for uid, data := range groupData {
		for bodyComId, v := range data {
			var romAver uint
			var degreeAver uint
			if len(v.romList) != 0 {
				romAver = sum(v.romList) / uint(len(v.romList))
			}
			if len(v.degreeList) != 0 {
				degreeAver = sum(v.degreeList) / uint(len(v.degreeList))
			}
			var clinicAver string

			// 빈도 수를 기록하기 위한 해시맵
			frequency := make(map[string]int)
			// 가장 많은 문자열과 그 빈도 수를 추적
			maxCount := 0

			// 각 문자열의 빈도 수를 해시맵에 기록하고 가장 많은 문자열을 찾기
			for _, str := range v.clinicList {
				frequency[str]++
				if frequency[str] > maxCount {
					clinicAver = str
					maxCount = frequency[str]
				}
			}

			// 해당 그룹의 사용자들에게 평균값을 설정
			var romAv *uint = nil
			if romAver != 0 {
				romAv = &romAver
			}

			var clinicalFeatureAv *string = nil
			if clinicAver != "" {
				clinicalFeatureAv = &clinicAver
			}
			var degreeAv *uint = nil
			if degreeAver != 0 {
				degreeAv = &degreeAver
			}
			userJointActionsMap[uid] = append(userJointActionsMap[uid], AfcResponse{
				BodyCompositionID: bodyComId,
				RomAv:             romAv,
				ClinicalFeatureAv: clinicalFeatureAv,
				DegreeAv:          degreeAv,
			})
		}

	}
	var response []SearchUserResponse
	for _, user := range users {
		ageCode := calculateAgeCode(user.Birthday)

		response = append(response, SearchUserResponse{
			ID:             user.ID,
			Name:           user.Name,
			Gender:         user.Gender,
			Phone:          user.Phone,
			AgeCode:        ageCode,
			RegistDay:      user.RegistDay.String(),
			AgencyId:       user.AdminID,
			AgencyName:     user.Agency.Name,
			AdminId:        user.AdminID,
			AdminName:      user.Admin.Name,
			UseStatusId:    user.UseStatusID,
			UseStatusName:  user.UseStatus.Name,
			DisableTypes:   disableTypeNamesMap[user.ID],
			VisitPurposes:  visitPurposeNamesMap[user.ID],
			DisableDetails: disableDetailNamesMap[user.ID],
			Afc:            userJointActionsMap[user.ID],
			Addr:           user.Addr + " " + user.AddrDetail,
			Birthday:       user.Birthday.Format("2006-01-02"),
			Memo:           user.Memo,
		})
	}

	return response, nil
}

func (service *adminService) getAgencis(id uint) ([]AgAdResponse, error) {
	var superAgencyID uint
	if err := service.db.Table("admins").
		Select("agencies.super_agency_id").
		Joins("JOIN agencies ON agencies.id = admins.agency_id").
		Where("admins.id = ?", id).
		Scan(&superAgencyID).Error; err != nil {
		return nil, errors.New("db error")
	}

	var agencis []model.Agency
	if err := service.db.Where("super_agency_id=?", superAgencyID).Find(&agencis).Error; err != nil {
		return nil, errors.New("db error")
	}

	var agencyResponse []AgAdResponse
	for _, agency := range agencis {
		agencyResponse = append(agencyResponse, AgAdResponse{ID: agency.ID, Name: agency.Name})
	}

	return agencyResponse, nil
}

func (service *adminService) getAdmins(adminId uint) ([]AgAdResponse, error) {

	// 해당 adminId를 기준으로 superAgencyID를 찾기
	var superAgencyID uint
	if err := service.db.Table("admins").
		Select("agencies.super_agency_id").
		Joins("JOIN agencies ON agencies.id = admins.agency_id").
		Where("admins.id = ?", adminId).
		Scan(&superAgencyID).Error; err != nil {
		return nil, errors.New("db error")
	}

	// superAgencyID를 기준으로 하위의 admin 조회
	var admins []model.Admin
	if err := service.db.Joins("JOIN agencies ON agencies.id = admins.agency_id").
		Where("agencies.super_agency_id = ?", superAgencyID).
		Find(&admins).Error; err != nil {
		return nil, errors.New("db error")
	}

	var agencyResponse []AgAdResponse
	for _, admin := range admins {
		agencyResponse = append(agencyResponse, AgAdResponse{ID: admin.ID, Name: admin.Name})
	}

	return agencyResponse, nil
}

func (service *adminService) getDisableDetails() ([]AgAdResponse, error) {
	var disableDetails []model.DisableDetail
	if err := service.db.Find(&disableDetails).Error; err != nil {
		return nil, errors.New("db error")
	}

	var agencyResponse []AgAdResponse
	for _, detail := range disableDetails {
		agencyResponse = append(agencyResponse, AgAdResponse{ID: detail.ID, Name: detail.Name})
	}

	return agencyResponse, nil
}

func (service *adminService) getAfcs(id, uid uint) (GetAfcResponse, error) {
	if result, err := checkFamily(service, id, uid); err != nil || !result {
		if !result {
			return GetAfcResponse{}, errors.New("not family")
		}
	}
	var afcs []model.UserAfc
	if err := service.db.Where("uid = ?", uid).Preload("Admin").Preload("UserAfcHistoryGroup.Admin").Find(&afcs).Error; err != nil {
		return GetAfcResponse{}, errors.New("db error")
	}

	var response GetAfcResponse
	for _, v := range afcs {
		response.UserAfcResponse = append(response.UserAfcResponse, UserAfcResponse{
			UpdatedAdmin:      v.Admin.Name,
			Updated:           v.CreatedAt.Format("2006-01-02 15:04:05"),
			BodyCompositionID: v.BodyCompositionID,
			JointActionID:     v.JointActionID,
			RomID:             v.RomID,
			ClinicalFeatureID: v.ClinicalFeatureID,
			DegreeID:          v.DegreeID,
		})
	}
	if len(afcs) != 0 {
		response.CreatedAdmin = afcs[0].UserAfcHistoryGroup.Admin.Name
		response.Created = afcs[0].UserAfcHistoryGroup.CreatedAt.Format("2006-01-02 15:04:05")
		response.GroupId = afcs[0].UserAfcHistoryGroupID
	}

	return response, nil
}

func (service *adminService) createAfc(request SaveAfcRequest) (string, error) {

	if result, err := checkFamily(service, request.Id, request.Uid); err != nil || !result {
		if !result {
			return "", errors.New("not family")
		}
	}
	if !validateAfc(request.Afcs) {
		return "", errors.New("validate afc")
	}

	tx := service.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	//history기록을 위해 기존의 afc불러옴
	var originAfcs []model.UserAfc
	if err := service.db.Where("uid=?", request.Uid).Find(&originAfcs).Error; err != nil {
		return "", errors.New("db error")
	}

	//히스토리 그룹생성
	var group model.UserAfcHistoryGroup
	group.AdminID = request.Id
	group.Uid = request.Uid
	if err := tx.Create(&group).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error1")
	}

	//UserAfc 생성
	var ujas []model.UserAfc
	for _, v := range request.Afcs {
		var joint *uint = nil
		var rom *uint = nil
		var clinic *uint = nil
		var degree *uint = nil

		if v.JointActionID != 0 {
			jointID := v.JointActionID
			joint = &jointID
		}

		if v.ClinicalFeatureID != uint(AC) {
			if v.RomID != 0 {
				romID := v.RomID
				rom = &romID
			}
		}

		if v.BodyCompositionID != uint(TR) && v.BodyCompositionID != uint(LOCOMOTION) {

			if v.ClinicalFeatureID != uint(AC) {
				if v.ClinicalFeatureID != 0 {
					clinicalFeatureID := v.ClinicalFeatureID // 새로운 변수를 생성하여 값을 복사합니다.
					clinic = &clinicalFeatureID
				}
				if v.DegreeID != 0 {
					degreeID := v.DegreeID // 새로운 변수를 생성하여 값을 복사합니다.
					degree = &degreeID
				}
			} else {
				if v.ClinicalFeatureID != 0 {
					clinicalFeatureID := v.ClinicalFeatureID // 새로운 변수를 생성하여 값을 복사합니다.
					clinic = &clinicalFeatureID
				}
			}
		}

		ujas = append(ujas, model.UserAfc{UserAfcHistoryGroupID: group.ID, AdminID: request.Id, Uid: request.Uid, BodyCompositionID: v.BodyCompositionID,
			JointActionID: joint, RomID: rom, ClinicalFeatureID: clinic, DegreeID: degree})
	}
	result := tx.Where("uid = ? ", request.Uid).Unscoped().Delete(&model.UserAfc{})
	if result.Error != nil {
		tx.Rollback()
		return "", errors.New("db error")
	}
	if err := tx.Create(&ujas).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}

	// 기존에 등록된게 있을때 히스토리 등록
	deletedRows := result.RowsAffected
	if deletedRows != 0 {
		var historis []model.UserAfcHistory
		for _, v := range originAfcs {
			historis = append(historis, model.UserAfcHistory{UserAfcHistoryGroupID: v.UserAfcHistoryGroupID, AdminID: request.Id,
				BodyCompositionID: v.BodyCompositionID, JointActionID: v.JointActionID, RomID: v.RomID, ClinicalFeatureID: v.ClinicalFeatureID, DegreeID: v.DegreeID})
		}
		if err := tx.Create(&historis).Error; err != nil {
			tx.Rollback()
			return "", errors.New("db error1")
		}
	}

	tx.Commit()
	return "200", nil
}

func (service *adminService) updateAfc(request SaveAfcRequest) (string, error) {
	if result, err := checkFamily(service, request.Id, request.Uid); err != nil || !result {
		if !result {
			return "", errors.New("not family")
		}
	}
	if !validateAfc(request.Afcs) {
		return "", errors.New("validate afc")
	}

	//기존의 히스토리그룹 id참조를 위해 afc 하나만가져옴
	var groupId uint
	row := service.db.Model(&model.UserAfc{}).Where("uid = ?", request.Uid).Select("user_afc_history_group_id").Row()
	if err := row.Scan(&groupId); err != nil {
		return "", errors.New("db error")
	}

	var ujas []model.UserAfc
	for _, v := range request.Afcs {
		var joint *uint = nil
		var rom *uint = nil
		var clinic *uint = nil
		var degree *uint = nil

		if v.JointActionID != 0 {
			jointID := v.JointActionID
			joint = &jointID
		}

		if v.ClinicalFeatureID != uint(AC) {
			if v.RomID != 0 {
				romID := v.RomID // 새로운 변수를 생성하여 값을 복사합니다.
				rom = &romID
			}
		}
		if v.BodyCompositionID != uint(TR) && v.BodyCompositionID != uint(LOCOMOTION) {
			if v.ClinicalFeatureID != uint(AC) {
				if v.ClinicalFeatureID != 0 {
					clinicalFeatureID := v.ClinicalFeatureID // 새로운 변수를 생성하여 값을 복사합니다.
					clinic = &clinicalFeatureID
				}
				if v.DegreeID != 0 {
					degreeID := v.DegreeID // 새로운 변수를 생성하여 값을 복사합니다.
					degree = &degreeID
				}
			} else {
				if v.ClinicalFeatureID != 0 {
					clinicalFeatureID := v.ClinicalFeatureID // 새로운 변수를 생성하여 값을 복사합니다.
					clinic = &clinicalFeatureID
				}
			}
		}

		ujas = append(ujas, model.UserAfc{UserAfcHistoryGroupID: groupId, AdminID: request.Id, Uid: request.Uid, BodyCompositionID: v.BodyCompositionID, JointActionID: joint,
			RomID: rom, ClinicalFeatureID: clinic, DegreeID: degree})
	}

	tx := service.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	result := tx.Where("uid = ? ", request.Uid).Unscoped().Delete(&model.UserAfc{})
	if result.Error != nil {
		tx.Rollback()
		return "", errors.New("db error")
	}

	if err := tx.Create(&ujas).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}

	tx.Commit()
	return "200", nil

}
func (service *adminService) getAfcHistoris(id, uid uint) ([]GetAfcResponse, error) {
	if result, err := checkFamily(service, id, uid); err != nil || !result {
		if !result {
			return nil, errors.New("not family")
		}
	}
	var groups []model.UserAfcHistoryGroup

	// Subquery to find UserAfcHistoryGroups not referenced by UserAfc
	subQuery := service.db.Model(&model.UserAfc{}).
		Select("user_afc_history_group_id").
		Where("uid = ?", uid)

	// Main query to find UserAfcHistoryGroups not in the subquery
	if err := service.db.
		Where("uid = ?", uid).
		Where("id NOT IN (?)", subQuery).Order("id DESC").Limit(2).
		Find(&groups).Error; err != nil {
		return nil, errors.New("db error")
	}

	groupIds := make([]uint, 0)
	for _, v := range groups {
		groupIds = append(groupIds, v.ID)
	}

	var afcs []model.UserAfcHistory
	if err := service.db.Where("user_afc_history_group_id IN ?", groupIds).Preload("UserAfcHistoryGroup.Admin").Preload("Admin").Find(&afcs).Error; err != nil {
		return nil, errors.New("db error2")
	}

	// Group by UserAfcHistoryGroupID
	historyGroupMap := make(map[uint]*GetAfcResponse)
	for _, v := range afcs {
		groupID := v.UserAfcHistoryGroupID
		if _, exists := historyGroupMap[groupID]; !exists {
			historyGroupMap[groupID] = &GetAfcResponse{
				CreatedAdmin:    v.UserAfcHistoryGroup.Admin.Name,
				Created:         v.UserAfcHistoryGroup.CreatedAt.Format("2006-01-02 15:04:05"),
				GroupId:         groupID,
				UserAfcResponse: []UserAfcResponse{},
			}
		}
		historyGroupMap[groupID].UserAfcResponse = append(historyGroupMap[groupID].UserAfcResponse, UserAfcResponse{
			UpdatedAdmin:      v.Admin.Name,
			Updated:           v.CreatedAt.Format("2006-01-02 15:04:05"),
			BodyCompositionID: v.BodyCompositionID,
			JointActionID:     v.JointActionID,
			RomID:             v.RomID,
			ClinicalFeatureID: v.ClinicalFeatureID,
			DegreeID:          v.DegreeID,
		})
	}

	// Convert map to slice
	var response []GetAfcResponse
	for _, v := range historyGroupMap {
		response = append(response, *v)
	}

	return response, nil
}

func (service *adminService) updateAfcHistory(request SaveAfcHistoryRequest) (string, error) {
	if !validateAfc(request.Afcs) {
		return "", errors.New("validate afc")
	}

	var group model.UserAfcHistoryGroup
	if err := service.db.Where("id =?", request.GroupId).First(&group).Error; err != nil {
		return "", errors.New("db error")
	}

	if result, err := checkFamily(service, request.Id, group.Uid); err != nil || !result {
		if !result {
			return "", errors.New("not family")
		}
	}

	var historis []model.UserAfcHistory

	for _, v := range request.Afcs {
		var joint *uint = nil
		var rom *uint = nil
		var clinic *uint = nil
		var degree *uint = nil

		if v.JointActionID != 0 {
			jointID := v.JointActionID
			joint = &jointID
		}

		if v.ClinicalFeatureID != uint(AC) {
			if v.RomID != 0 {
				romID := v.RomID
				rom = &romID
			}
		}

		if v.BodyCompositionID != uint(TR) && v.BodyCompositionID != uint(LOCOMOTION) {

			if v.ClinicalFeatureID != uint(AC) {
				if v.ClinicalFeatureID != 0 {
					clinicalFeatureID := v.ClinicalFeatureID // 새로운 변수를 생성하여 값을 복사합니다.
					clinic = &clinicalFeatureID
				}
				if v.DegreeID != 0 {
					degreeID := v.DegreeID // 새로운 변수를 생성하여 값을 복사합니다.
					degree = &degreeID
				}
			} else {
				if v.ClinicalFeatureID != 0 {
					clinicalFeatureID := v.ClinicalFeatureID // 새로운 변수를 생성하여 값을 복사합니다.
					clinic = &clinicalFeatureID
				}
			}
		}
		historis = append(historis, model.UserAfcHistory{UserAfcHistoryGroupID: request.GroupId, AdminID: request.Id, BodyCompositionID: v.BodyCompositionID, JointActionID: joint, RomID: rom, ClinicalFeatureID: clinic, DegreeID: degree})
	}

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	result := tx.Where("user_afc_history_group_id = ? ", request.GroupId).Unscoped().Delete(&model.UserAfcHistory{})
	if result.Error != nil {
		tx.Rollback()
		return "", errors.New("db error")
	}

	if err := tx.Create(&historis).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}

	tx.Commit()
	return "200", nil
}

func (service *adminService) searchDiary(request SearchDiaryRequest) ([]SearchDiaryResponse, error) {
	var response []SearchDiaryResponse

	pageSize := 10
	offset := int(request.Page) * pageSize

	var diaris []model.Diary
	query := service.db.Model(&model.Diary{})

	if strings.TrimSpace(request.Name) != "" {
		query = query.Where("name LIKE ?", "%"+request.Name+"%")
	}
	if request.AdminID != 0 {
		query = query.Where("admin_id = ?", request.AdminID)
	}
	if request.ClassType != 0 {
		query = query.Where("class_type = ?", request.ClassType)
	}
	if len(request.DisableTypeIDs) > 0 {
		query = query.Where("id IN (SELECT uid FROM user_disables WHERE disable_type_id IN ?)", request.DisableTypeIDs)
	}
	if len(request.VisitPurposeIDs) > 0 {
		query = query.Where("id IN (SELECT uid FROM user_visits WHERE visit_purpose_id IN ?)", request.VisitPurposeIDs)
	}
	if len(request.DisableDetailIDs) > 0 {
		query = query.Where("id IN (SELECT uid FROM user_disable_details WHERE disable_detail_id IN ?)", request.DisableDetailIDs)
	}

	if len(request.ClassPurposeIDs) > 0 {
		query = query.Where("id IN (SELECT diary_id FROM diary_class_purposes WHERE class_purpose_id IN ?)", request.ClassPurposeIDs)
	}

	if strings.TrimSpace(request.ClassDate) != "" {
		classDate, err := time.Parse("2006-01-02", request.ClassDate)
		if err != nil {
			return nil, errors.New("invalid date format for ClassDate")
		}
		query = query.Where("class_date = ?", classDate)
	}

	query = query.Offset(offset).Limit(pageSize)
	if err := query.Preload("User.Admin").Find(&diaris).Error; err != nil {
		return nil, err
	}

	// 다이어리 ID 리스트 가져오기
	diaryIDs := make([]uint, len(diaris))
	for i, dairy := range diaris {
		diaryIDs[i] = dairy.ID
	}

	var puposes []model.DiaryClassPurpose
	if err := service.db.Where("diary_id IN ?", diaryIDs).Preload("ClassPurpose").Find(&puposes).Error; err != nil {
		return nil, err
	}

	var exerciseDiarys []model.ExerciseDiary
	if err := service.db.Where("diary_id IN ?", diaryIDs).Preload("Exercise").Preload("Measure").Find(&exerciseDiarys).Error; err != nil {
		return nil, err
	}

	edrs := make(map[uint]map[uint]*ExerciseDiaryResponse)
	for _, v := range exerciseDiarys {
		if _, ok := edrs[v.DiaryID]; !ok {
			edrs[v.DiaryID] = make(map[uint]*ExerciseDiaryResponse)
		}
		if _, ok := edrs[v.DiaryID][v.ExerciseID]; !ok {
			edrs[v.DiaryID][v.ExerciseID] = &ExerciseDiaryResponse{
				ExerciseID:   v.ExerciseID,
				ExerciseName: v.Exercise.Name,
				Measures:     []MeasureResponse{},
			}
		}
		edrs[v.DiaryID][v.ExerciseID].Measures = append(edrs[v.DiaryID][v.ExerciseID].Measures, MeasureResponse{
			MeasureID:   v.MeasureID,
			MeasureName: v.Measure.Name,
			Value:       v.Value,
		})

	}
	purposeMap := make(map[uint][]IdNameResponse)
	for _, v := range puposes {
		purposeMap[v.DiaryID] = append(purposeMap[v.DiaryID], IdNameResponse{Id: v.ClassPurposeID, Name: v.ClassPurpose.Name})
	}

	for _, v := range diaris {
		var quillJson []QuillJson
		if err := json.Unmarshal(v.Explain, &quillJson); err != nil {
			return nil, err
		}

		// ExerciseDiaryResponse 맵을 슬라이스로 변환
		exerciseMeasures := make([]ExerciseDiaryResponse, 0)
		if exerciseMap, ok := edrs[v.ID]; ok {
			for _, edr := range exerciseMap {
				exerciseMeasures = append(exerciseMeasures, *edr)
			}
		}
		response = append(response, SearchDiaryResponse{ID: v.ID, CreatedAt: v.CreatedAt.Format("2006-01-02"), UpdatedAt: v.UpdatedAt.Format("2006-01-02"), Uid: v.Uid, UserName: v.User.Name,
			DiaryName: v.Title, ClassName: v.ClassName, ClassType: v.ClassType, ClassDate: v.ClassDate.Format("2006-01-02"), AdminName: v.User.Admin.Name, Explain: quillJson,
			ClassPurposes: purposeMap[v.ID], ExerciseMeasures: exerciseMeasures})
	}

	return response, nil
}

func (service *adminService) saveDiary(request SaveDiaryRequest) (string, error) {
	if !validateDiary(request) {
		return "", errors.New("validate field")
	}

	if result, err := checkFamily(service, request.AdminId, request.Uid); err != nil || !result {
		if !result {
			return "", errors.New("not family")
		}
	}

	classDate, err := time.Parse("2006-01-02", request.ClassDate)
	if err != nil {
		return "", errors.New("validate date")
	}

	for i, v := range request.Explain {
		switch insertValue := v.Insert.(type) {
		case map[string]interface{}:
			if image, ok := insertValue["image"]; ok {
				base64Image, ok := image.(string)
				if !ok {
					return "", errors.New("image field is not a string")
				}
				imgData, err := base64.StdEncoding.DecodeString(base64Image)
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
				url, err := uploadImagesToS3(imgData, contentType, ext, service.s3svc, service.bucket, service.bucketUrl, strconv.FormatUint(uint64(request.Uid), 10))
				if err != nil {
					return "", err
				}
				request.Explain[i].Insert = map[string]interface{}{"image": url}
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

	tx := service.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	var diary model.Diary
	if request.Id != 0 {
		// ID가 주어졌을 때 기존 다이어리를 찾습니다.
		if err := tx.First(&diary, request.Id).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				return "", errors.New("db error")
			}
		}
	}
	diary.Title = request.Title
	diary.Uid = request.Uid
	diary.ClassDate = classDate
	diary.ClassName = request.ClassName
	diary.ClassType = request.ClassType
	diary.AdminID = request.AdminId
	diary.Explain = explainJson

	if err := tx.Save(&diary).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error1")
	}

	var diaryClassPurposes []model.DiaryClassPurpose
	for _, v := range request.ClassPurposeIDs {
		diaryClassPurposes = append(diaryClassPurposes, model.DiaryClassPurpose{ClassPurposeID: v, DiaryID: diary.ID})
	}

	if err := tx.Where("diary_id = ?", diary.ID).Delete(&model.DiaryClassPurpose{}).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}
	if err := tx.Create(&diaryClassPurposes).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error3")
	}

	var exerciseDairis []model.ExerciseDiary
	for _, v := range request.ExerciseMeasures {
		for _, measure := range v.Measures {
			exerciseDairis = append(exerciseDairis, model.ExerciseDiary{ExerciseID: v.ExerciseID, MeasureID: measure.MeasureID, DiaryID: diary.ID, Value: measure.Value})
		}
	}

	if err := tx.Where("diary_id = ?", diary.ID).Delete(&model.ExerciseDiary{}).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error4")
	}

	if err := tx.Create(&exerciseDairis).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error5")
	}

	tx.Commit()
	return "200", nil
}
