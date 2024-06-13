// /user-service/service/service.go

package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"

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
