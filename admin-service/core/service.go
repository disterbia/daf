// /user-service/service/service.go

package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"admin-service/model"
	pb "admin-service/proto"

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
	getAgencis() ([]AgAdResponse, error)
	getAdmins() ([]AgAdResponse, error)
	getDisableDetails() ([]AgAdResponse, error)
	getAfcs(uid uint) ([]GetAfcResponse, error)
	createAfc(request SaveAfcRequest) (string, error)
	updateAfc(request SaveAfcRequest) (string, error)
	getAfcHistoris(uid uint) ([]GetAfcResponse, error)
	updateAfcHistory(request SaveAfcHistoryRequest) (string, error)
}

type adminService struct {
	db          *gorm.DB
	emailClient pb.EmailServiceClient
}

func NewAdminService(db *gorm.DB, conn *grpc.ClientConn) AdminService {
	emailClient := pb.NewEmailServiceClient(conn)
	return &adminService{db: db, emailClient: emailClient}
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
	pageSize := 10
	offset := int(request.Page) * pageSize

	var users []model.User
	query := service.db.Model(&model.User{})

	if strings.TrimSpace(request.Name) != "" {
		query = query.Where("name = ?", request.Name)
	}
	if request.Gender != 0 {
		query = query.Where("gender = ?", request.Gender)
	}
	if request.AgencyID != 0 {
		query = query.Where("agency_id = ?", request.AgencyID)
	}
	if request.AdminID != 0 {
		query = query.Where("admin_id = ?", request.AdminID)
	}
	if request.UseStatusID != 0 {
		query = query.Where("use_status_id = ?", request.UseStatusID)
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
	// Afc 조건 추가
	if len(request.Afcs) > 0 {
		for _, afc := range request.Afcs {
			if afc.BodyCompositionID != 0 {
				query = query.Where("id IN (SELECT uid FROM user_joint_actions WHERE body_composition_id = ?)", afc.BodyCompositionID)
			}
			if afc.JointActionID != 0 {
				query = query.Where("id IN (SELECT uid FROM user_joint_actions WHERE joint_action_id = ?)", afc.JointActionID)
			}
			if afc.RomID != 0 {
				query = query.Where("id IN (SELECT uid FROM user_joint_actions WHERE rom_id = ?)", afc.RomID)
			}
			if afc.ClinicalFeatureID != 0 {
				query = query.Where("id IN (SELECT uid FROM user_joint_actions WHERE clinical_feature_id = ?)", afc.ClinicalFeatureID)
			}
			if afc.DegreeID != 0 {
				query = query.Where("id IN (SELECT uid FROM user_joint_actions WHERE degree_id = ?)", afc.DegreeID)
			}
		}
	}

	if request.AgeCode != 0 {
		startDate, endDate, err := getBirthdayRangeByAgeCode(request.AgeCode)
		if err != nil {
			return nil, err
		}
		query = query.Where("birthday BETWEEN ? AND ?", startDate, endDate)
	}

	if strings.TrimSpace(request.RegistDay) != "" {
		registDay, err := time.Parse("2006-01-02", request.RegistDay)
		if err != nil {
			return nil, errors.New("invalid date format for registDay")
		}
		query = query.Where("regist_day = ?", registDay)
	}

	query = query.Offset(offset).Limit(pageSize)

	if err := query.Preload("Agency").Preload("Admin").Preload("UserStatus").Find(&users).Error; err != nil {
		return nil, err
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
	var userJointActions []model.UserAfc

	service.db.Where("uid IN ?", userIDs).Preload("DisableType").Find(&userDisables)
	service.db.Where("uid IN ?", userIDs).Preload("VisitPurpose").Find(&userVisits)
	service.db.Where("uid IN ?", userIDs).Preload("DisableDetail").Find(&userDisableDetails)
	service.db.Where("uid IN ?", userIDs).Find(&userJointActions)
	// 사용자 ID를 키로 하는 맵 생성
	disableTypeNamesMap := make(map[uint][]string)
	visitPurposeNamesMap := make(map[uint][]string)
	disableDetailNamesMap := make(map[uint][]string)
	userJointActionsMap := make(map[uint][]AfcResponse)
	type GroupData struct {
		romList    uint
		clinicList []string
		degreeList uint
		count      uint
	}
	groupData := make(map[uint]GroupData)

	for _, ud := range userDisables {
		disableTypeNamesMap[ud.UID] = append(disableTypeNamesMap[ud.UID], ud.DisableType.Name)
	}

	for _, uv := range userVisits {
		visitPurposeNamesMap[uv.UID] = append(visitPurposeNamesMap[uv.UID], uv.VisitPurpose.Name)
	}

	for _, udd := range userDisableDetails {
		disableDetailNamesMap[udd.UID] = append(disableDetailNamesMap[udd.UID], udd.DisableDetail.Name)
	}

	for _, uja := range userJointActions {
		// BodyCompositionID를 키로 사용하는 그룹에 데이터 추가
		bodyCompId := uja.BodyCompositionID
		data := groupData[bodyCompId]
		data.romList += uja.RomID
		if uja.ClinicalFeatureID != nil && uja.DegreeID != nil {
			data.clinicList = append(data.clinicList, uja.ClinicalFeature.Code)
			data.degreeList += *uja.DegreeID
		}
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
		// 해당 그룹의 사용자들에게 평균값을 설정
		for _, uja := range userJointActions {
			if uja.BodyCompositionID == bodyCompId {
				romName := new(string)
				if bodyCompId == uint(LOCOMOTION) {
					romName = uja.Rom.Name
				}
				clinicalFeatureAv := new(string)
				if clinicAver != "" {
					*clinicalFeatureAv = clinicAver
				}
				degreeAv := new(uint)
				if degreeAver != 0 {
					*degreeAv = degreeAver
				}
				userJointActionsMap[uja.Uid] = append(userJointActionsMap[uja.Uid], AfcResponse{
					BodyCompositionID: uja.BodyCompositionID,
					RomAv:             romAver,
					RomName:           romName,
					ClinicalFeatureAv: clinicalFeatureAv,
					DegreeAv:          degreeAv,
				})
			}
		}
	}
	var response []SearchUserResponse
	for _, user := range users {
		ageCode := calculateAgeCode(user.Birthday)

		response = append(response, SearchUserResponse{
			ID:                 user.ID,
			Name:               user.Name,
			Gender:             user.Gender,
			AgeCode:            ageCode,
			RegistDay:          user.RegistDay.String(),
			AgencyName:         user.Agency.Name,
			AdminName:          user.Admin.Name,
			UseStatusName:      user.UseStatus.Name,
			DisableTypeNames:   disableTypeNamesMap[user.ID],
			VisitPurposeNames:  visitPurposeNamesMap[user.ID],
			DisableDetailNames: disableDetailNamesMap[user.ID],
			Afc:                userJointActionsMap[user.ID],
		})
	}

	return response, nil
}

func (service *adminService) getAgencis() ([]AgAdResponse, error) {
	var agencis []model.Agency
	if err := service.db.Find(&agencis).Error; err != nil {
		return nil, errors.New("db error")
	}

	var agencyResponse []AgAdResponse
	for _, agency := range agencis {
		agencyResponse = append(agencyResponse, AgAdResponse{ID: agency.ID, Name: agency.Name})
	}

	return agencyResponse, nil
}

func (service *adminService) getAdmins() ([]AgAdResponse, error) {
	var admins []model.Admin
	if err := service.db.Find(&admins).Error; err != nil {
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

func (service *adminService) getAfcs(uid uint) ([]GetAfcResponse, error) {
	var afcs []model.UserAfc
	if err := service.db.Where("uid = ?", uid).Preload("Admin").Preload("UserAfcHistoryGroup.Admin").Find(&afcs).Error; err != nil {
		return nil, errors.New("db error")
	}

	// Group by UserAfcHistoryGroupID
	historyGroupMap := make(map[uint]*GetAfcResponse)
	for _, v := range afcs {
		groupID := v.UserAfcHistoryGroupID
		if _, exists := historyGroupMap[groupID]; !exists {
			historyGroupMap[groupID] = &GetAfcResponse{
				CreatedAdmin:    v.UserAfcHistoryGroup.Admin.Name,
				Created:         v.UserAfcHistoryGroup.CreatedAt.Format("2006-01-02 15:04:05"),
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

func (service *adminService) createAfc(request SaveAfcRequest) (string, error) {
	if !validateAfc(request.Afcs) {
		return "", errors.New("validate afc")
	}
	var ujas []model.UserAfc

	for _, v := range request.Afcs {
		clinic := new(uint)
		degree := new(uint)
		if v.BodyCompositionID != uint(TR) && v.BodyCompositionID != uint(LOCOMOTION) {
			clinic = &v.ClinicalFeatureID
			degree = &v.DegreeID
		}
		ujas = append(ujas, model.UserAfc{AdminID: request.Id, Uid: request.Uid, BodyCompositionID: v.BodyCompositionID, JointActionID: v.JointActionID, RomID: v.RomID, ClinicalFeatureID: clinic, DegreeID: degree})
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
	var group model.UserAfcHistoryGroup
	group.AdminID = request.Id
	group.Uid = request.Uid
	if err := tx.Create(&group).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error0")
	}
	deletedRows := result.RowsAffected
	if deletedRows != 0 { // 기존에 등록된게 있을때
		var historis []model.UserAfcHistory
		for _, v := range request.Afcs {
			clinic := new(uint)
			degree := new(uint)
			if v.BodyCompositionID != uint(TR) && v.BodyCompositionID != uint(LOCOMOTION) {
				clinic = &v.ClinicalFeatureID
				degree = &v.DegreeID
			}
			historis = append(historis, model.UserAfcHistory{UserAfcHistoryGroupID: group.ID, AdminID: request.Id,
				BodyCompositionID: v.BodyCompositionID, JointActionID: v.JointActionID, RomID: v.RomID, ClinicalFeatureID: clinic, DegreeID: degree})
		}
		if err := tx.Create(&historis).Error; err != nil {
			tx.Rollback()
			return "", errors.New("db error1")
		}
	}

	if err := tx.Create(&ujas).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}

	tx.Commit()
	return "200", nil
}

func (service *adminService) updateAfc(request SaveAfcRequest) (string, error) {
	if !validateAfc(request.Afcs) {
		return "", errors.New("validate afc")
	}
	var ujas []model.UserAfc

	for _, v := range request.Afcs {
		clinic := new(uint)
		degree := new(uint)
		if v.BodyCompositionID != uint(TR) && v.BodyCompositionID != uint(LOCOMOTION) {
			clinic = &v.ClinicalFeatureID
			degree = &v.DegreeID
		}
		ujas = append(ujas, model.UserAfc{AdminID: request.Id, Uid: request.Uid, BodyCompositionID: v.BodyCompositionID, JointActionID: v.JointActionID, RomID: v.RomID, ClinicalFeatureID: clinic, DegreeID: degree})
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
func (service *adminService) getAfcHistoris(uid uint) ([]GetAfcResponse, error) {
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

	var historis []model.UserAfcHistory

	for _, v := range request.Afcs {
		clinic := new(uint)
		degree := new(uint)
		if v.BodyCompositionID != uint(TR) && v.BodyCompositionID != uint(LOCOMOTION) {
			clinic = &v.ClinicalFeatureID
			degree = &v.DegreeID
		}
		historis = append(historis, model.UserAfcHistory{AdminID: request.Id, BodyCompositionID: v.BodyCompositionID, JointActionID: v.JointActionID, RomID: v.RomID, ClinicalFeatureID: clinic, DegreeID: degree})
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
