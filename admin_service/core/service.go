// /user-service/service/service.go

package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"admin_service/model"
	pb "admin_service/proto"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type AdminService interface {
	sendAuthCode(number string) (string, error)
	login(request LoginRequest) (string, error)
	verifyAuthCode(number, code string) (string, error)
	signIn(request SignInRequest) (string, error)
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
	if request.Password == "" {
		return "", errors.New("empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	if err := service.db.Debug().Where("email=? AND password=?", request.Email, string(hashedPassword)).First(&u).Error; err != nil {
		return "", err
	}

	// 새로운 JWT 토큰 생성
	tokenString, err := generateJWT(u)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (service *adminService) sendAuthCode(email string) (string, error) {

	err := validateEmail(email)
	if err != nil {
		return "", err
	}

	result := service.db.Debug().Where("email=?", email).Find(&model.Admin{})
	if result.Error != nil {
		return "", errors.New("db error")

	} else if result.RowsAffected > 0 {
		// 레코드가 존재할 때
		return "", errors.New("-1")
	}

	var sb strings.Builder
	for i := 0; i < 6; i++ {
		fmt.Fprintf(&sb, "%d", rand.Intn(10)) // 0부터 9까지의 숫자를 무작위로 선택
	}

	go func() {
		reponse, err := service.emailClient.SendEmail(context.Background(), &pb.EmailRequest{
			Email: email, // 받는 사람의 이메일
			Code:  sb.String(),
		})
		if err != nil {
			log.Printf("Failed to send email: %v", err)
		}
		log.Printf(" send email: %v", reponse)
	}()

	return "200", nil
}

func (service *adminService) verifyAuthCode(email, code string) (string, error) {
	var authCode model.AuthCode

	if err := service.db.Where("email = ? ", email).Last(&authCode).Error; err != nil {
		return "", errors.New("db error")
	}
	if authCode.Code != code {
		return "", errors.New("-1")
	}
	if err := service.db.Create(&model.VerifiedEmail{Email: authCode.Email}).Error; err != nil {
		return "", errors.New("db error2")
	}

	return "200", nil
}

func (service *adminService) signIn(request SignInRequest) (string, error) {

	result := service.db.Debug().Where("email=?", request.Email).Find(&model.VerifiedEmail{})
	if result.Error != nil {
		return "", errors.New("db error")

	} else if result.RowsAffected > 0 {
		// 레코드가 존재할 때
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

	admin.Password = string(hashedPassword)
	if err := service.db.Create(&admin).Error; err != nil {
		return "", errors.New("db error")
	}
	return "200", nil
}
