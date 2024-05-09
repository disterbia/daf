// /user-service/service/service.go

package core

import (
	"encoding/base64"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"user_service/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
)

type UserService interface {
	snsLogin(request LoginRequest) (string, error)
	autoLogin(request AutoLoginRequest) (string, error)
	getUser(id uint) (UserResponse, error) //유저조회
	setUser(userRequest UserRequest) (string, error)
	removeUser(id uint) (string, error)
	getVersion() (AppVersionResponse, error)
	removeProfile(uid uint) (string, error)
}

type userService struct {
	db        *gorm.DB
	s3svc     *s3.S3
	bucket    string
	bucketUrl string
}

func NewUserService(db *gorm.DB, s3svc *s3.S3, bucket string, bucketUrl string) UserService {
	return &userService{db: db, s3svc: s3svc, bucket: bucket, bucketUrl: bucketUrl}
}

type PublicKey struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []PublicKey `json:"keys"`
}

func (service *userService) snsLogin(request LoginRequest) (string, error) {
	if request.FCMToken == "" || request.DeviceID == "" {
		return "", errors.New("check fcm_token,device_id")
	}
	iss := decodeJwt(request.IdToken)
	var user model.User
	var err error
	var email string
	var snsType uint

	if strings.Contains(iss, "kakao") { // 카카오
		snsType = uint(Kakao)
		if email, err = kakaoLogin(request); err != nil {
			return "", err
		}
	} else if strings.Contains(iss, "google") { // 구글
		snsType = uint(Google)
		if email, err = googleLogin(request); err != nil {
			return "", err
		}
	} else if strings.Contains(iss, "apple") { // 애플
		snsType = uint(Apple)
		if email, err = appleLogin(request); err != nil {
			return "", err
		}
	}

	if err := copyStruct(request, &user); err != nil {
		return "", err
	}

	user.Email = email
	user.SnsType = snsType
	u, err := findOrCreateUser(user, service)
	if err != nil {
		return "", err
	}

	// JWT 토큰 생성
	tokenString, err := generateJWT(u)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (service *userService) autoLogin(request AutoLoginRequest) (string, error) {
	if request.FcmToken == "" || request.DeviceId == "" {
		return "", errors.New("check fcm_token,device_id")
	}
	// 데이터베이스에서 사용자 조회
	var u model.User
	if err := service.db.Where("email = ?", request.Email).First(&u).Error; err != nil {
		return "", errors.New("db error")
	}
	// 새로운 JWT 토큰 생성
	tokenString, err := generateJWT(u)
	if err != nil {
		return "", err
	}

	if err := service.db.Model(&u).Updates(model.User{FCMToken: request.FcmToken, DeviceID: request.DeviceId}).Error; err != nil {
		return "", errors.New("db error2")
	}
	return tokenString, nil
}

func (service *userService) getUser(id uint) (UserResponse, error) {
	var user model.User
	result := service.db.Debug().Preload("Images", "type = ?", profileImageType).First(&user, id)
	if result.Error != nil {
		return UserResponse{}, errors.New("db error")
	}
	log.Println(user.CreatedAt)
	var userResponse UserResponse
	if err := copyStruct(user, &userResponse); err != nil {
		return UserResponse{}, err
	}

	if len(user.Images) != 0 {
		urlkey := extractKeyFromUrl(user.Images[0].Url, service.bucket, service.bucketUrl)
		thumbnailUrlkey := extractKeyFromUrl(user.Images[0].ThumbnailUrl, service.bucket, service.bucketUrl)
		// 사전 서명된 URL을 생성
		url, _ := service.s3svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(service.bucket),
			Key:    aws.String(urlkey),
		})
		thumbnailUrl, _ := service.s3svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(service.bucket),
			Key:    aws.String(thumbnailUrlkey),
		})
		urlStr, err := url.Presign(5 * time.Second) // URL은 5초 동안 유효
		if err != nil {
			return UserResponse{}, err
		}
		thumbnailUrlStr, err := thumbnailUrl.Presign(5 * time.Second) // URL은 5초 동안 유효 CachedNetworkImage 에서 캐싱해서 쓰면됨
		if err != nil {
			return UserResponse{}, err
		}
		userResponse.ProfileImage.Url = urlStr // 사전 서명된 URL로 업데이트
		userResponse.ProfileImage.ThumbnailUrl = thumbnailUrlStr

	}

	return userResponse, nil
}

func (service *userService) setUser(userRequest UserRequest) (string, error) {

	var fileName, thumbnailFileName string
	var image model.Image
	var user model.User

	user.ID = userRequest.ID
	user.Nickname = userRequest.Nickname

	if userRequest.ProfileImage != "" {

		//base64 string decode
		imgData, err := base64.StdEncoding.DecodeString(userRequest.ProfileImage)
		if err != nil {
			return "", err
		}

		//이미지 포맷 체크
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

		// 썸네일 이미지 생성
		thumbnailData, err := createThumbnail(imgData)
		if err != nil {
			return "", err
		}

		// S3에 이미지 및 썸네일 업로드
		fileName, thumbnailFileName, err = uploadImagesToS3(imgData, thumbnailData, contentType, ext, service.s3svc, service.bucket, service.bucketUrl, strconv.FormatUint(uint64(user.ID), 10))
		if err != nil {

			return "", err
		}
		image = model.Image{
			Uid:          user.ID,
			Url:          fileName,
			ThumbnailUrl: thumbnailFileName,
			ParentId:     user.ID,
			Type:         uint(profileImageType),
		}
	}

	// 트랜잭션 시작
	tx := service.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	//유저 정보 업데이트
	result := service.db.Model(&user).Where("id=?", user.ID).Update("nickname", user.Nickname)
	if result.Error != nil {
		log.Println(result.Error.Error())
		tx.Rollback()

		// 이미 업로드된 파일들을 S3에서 삭제

		if userRequest.ProfileImage != "" {
			go func() {
				deleteFromS3(fileName, service.s3svc, service.bucket, service.bucketUrl)
				deleteFromS3(thumbnailFileName, service.s3svc, service.bucket, service.bucketUrl)
			}()
		}

		return "", errors.New("db error")
	}
	if userRequest.ProfileImage != "" {
		// 기존 이미지 레코드 논리삭제
		result = service.db.Where("parent_id = ? AND type =?", user.ID, profileImageType).Delete(&model.Image{})
		if result.Error != nil {
			log.Println(result.Error.Error())
			tx.Rollback()
			if userRequest.ProfileImage != "" {
				go func() {
					deleteFromS3(fileName, service.s3svc, service.bucket, service.bucketUrl)
					deleteFromS3(thumbnailFileName, service.s3svc, service.bucket, service.bucketUrl)
				}()
			}
			return "", errors.New("db error4")
		}
		// 이미지 레코드 재 생성

		if err := tx.Create(&image).Error; err != nil {
			log.Println(err)
			tx.Rollback()
			if userRequest.ProfileImage != "" {
				go func() {
					deleteFromS3(fileName, service.s3svc, service.bucket, service.bucketUrl)
					deleteFromS3(thumbnailFileName, service.s3svc, service.bucket, service.bucketUrl)
				}()
			}
			return "", errors.New("db error5")
		}
	}

	tx.Commit()
	return "200", nil
}

func (service *userService) removeUser(id uint) (string, error) {
	var user model.User
	user.ID = id
	if err := service.db.Delete(&user).Error; err != nil {
		return "", errors.New("db error")
	}
	return "200", nil
}

func (service *userService) getVersion() (AppVersionResponse, error) {
	var version model.AppVersion
	result := service.db.Last(&version)
	if result.Error != nil {
		return AppVersionResponse{}, errors.New("db error")
	}
	var versionResponse AppVersionResponse
	if err := copyStruct(version, &versionResponse); err != nil {
		return AppVersionResponse{}, err
	}
	return versionResponse, nil
}

func (service *userService) removeProfile(uid uint) (string, error) {

	// 기존 이미지 레코드 논리삭제
	result := service.db.Where("parent_id = ? AND type =?", uid, profileImageType).Delete(&model.Image{})
	if result.Error != nil {
		return "", errors.New("db error2")
	}
	return "200", nil
}
