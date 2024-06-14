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

	if err := tx.Where("email = ?", request.Email).Unscoped().Delete(&verify).Error; err != nil {
		tx.Rollback()
		return "", errors.New("db error2")
	}

	admin.Password = string(hashedPassword)
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
		return "", errors.New("duplicate Id")
	}
	if checkDuplicates(request.VisitPurposeIDs) {
		return "", errors.New("duplicate Id")
	}
	if checkDuplicates(request.DisableDetailIDs) {
		return "", errors.New("duplicate Id")
	}
	birthday, err := time.Parse("2006-01-02", request.Birthday)
	if err != nil {
		return "", errors.New("date err")
	}
	registday, err := time.Parse("2006-01-02", request.RegistDay)
	if err != nil {
		return "", errors.New("date err")
	}

	var vists []model.UserVisit
	var disables []model.UserDisable
	var details []model.UserDisableDetail
	var user model.User

	tx := service.db.Begin()

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
