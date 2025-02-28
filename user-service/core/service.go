// /user-service/service/service.go

package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"user-service/model"
	pb "user-service/proto"

	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
)

type UserService interface {
	appleLogin(code string) (string, error)
	googleLogin(code string) (string, error)
	kakaoLogin(code string) (string, error)
	facebookLogin(code string) (string, error)
	naverLogin(code string) (string, error)

	// snsLogin(request LoginRequest) (string, error)
	// getUser(id uint) (UserResponse, error) //유저조회
	// setUser(userRequest UserRequest) (string, error)
	// removeUser(id uint) (string, error)
	// getVersion() (AppVersionResponse, error)
	// removeProfile(uid uint) (string, error)
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

type UserServer struct {
	pb.UnimplementedUserServiceServer
	DB        *gorm.DB
	S3svc     *s3.S3
	Bucket    string
	BucketUrl string
}

func (s *userService) appleLogin(code string) (string, error) {
	var user model.User
	var err error
	var email string

	clientID := os.Getenv("APPLE_CLIENT_ID")
	keyID := os.Getenv("APPLE_KEY_ID")
	teamID := os.Getenv("APPLE_TEAM_ID")
	privateKey := os.Getenv("APPLE_PRIVATE_KEY")

	// client_secret 생성
	clientSecret, err := GenerateClientSecret(keyID, teamID, clientID, privateKey)
	if err != nil {
		return "", err
	}

	// 요청 데이터 생성
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", "https://haruharu-daf.com/user/apple/callback")

	// POST 요청 생성
	req, err := http.NewRequest("POST", "https://appleid.apple.com/auth/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 응답 확인
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 응답 파싱
	var tokenResponse AppleTokenResponse
	tokenResponse.Code = code
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}
	if email, err = appleLogin(tokenResponse.IDToken); err != nil {
		return "", err
	}

	user.Email = &email
	user.SnsType = uint(Apple)
	user.UserType = uint(ADAPFIT)
	u, err := findOrCreateUser(user, s)
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

func (s *userService) googleLogin(code string) (string, error) {
	var user model.User
	var err error
	var email string

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := "http://haruharu-daf.com/user/google/callback"

	// 요청 데이터 생성
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)

	// POST 요청 생성
	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 응답 확인
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 응답 파싱
	var tokenResponse GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	// ID 토큰 검증 및 이메일 추출
	if email, err = googleLogin(tokenResponse.IDToken, clientID); err != nil {
		return "", err
	}

	// 유저 데이터 처리
	user.Email = &email
	user.SnsType = uint(Google)
	user.UserType = uint(ADAPFIT)
	u, err := findOrCreateUser(user, s)
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

func (s *userService) kakaoLogin(code string) (string, error) {
	var user model.User
	var err error
	var email string

	clientID := os.Getenv("KAKAO_CLIENT_ID")
	clientSecret := os.Getenv("KAKAO_CLIENT_SECRET")
	redirectURI := "https://localhost:44403/kakao/callback"

	// 요청 데이터 생성
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)

	// POST 요청 생성
	req, err := http.NewRequest("POST", "https://kauth.kakao.com/oauth/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 응답 확인
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 응답 파싱
	var tokenResponse KakaoTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	// ID 토큰 검증 및 이메일 추출
	if email, err = kakaoLogin(tokenResponse.IDToken); err != nil {
		return "", err
	}

	// 유저 데이터 처리
	user.Email = &email
	user.SnsType = uint(Kakao) // Kakao 상수는 별도로 정의 필요 (예: const Kakao = 3)
	user.UserType = uint(ADAPFIT)
	u, err := findOrCreateUser(user, s)
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

func (s *userService) facebookLogin(code string) (string, error) {
	var user model.User
	var err error
	var email string

	clientID := os.Getenv("FACEBOOK_CLIENT_ID")
	clientSecret := os.Getenv("FACEBOOK_CLIENT_SECRET")
	redirectURI := "http://localhost:44403/facebook/callback"

	// Access Token 요청
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("code", code)

	req, err := http.NewRequest("POST", "https://graph.facebook.com/v18.0/oauth/access_token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 응답 JSON을 구조체로 변환
	var tokenResponse FacebookTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	// 액세스 토큰이 없으면 에러 반환
	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("failed to retrieve access token from Facebook")
	}

	// 페이스북 사용자 정보 요청
	email, err = getFacebookUserInfo(tokenResponse.AccessToken)
	if err != nil {
		return "", err
	}

	// 유저 데이터 처리
	user.Email = &email
	user.SnsType = uint(Facebook)
	user.UserType = uint(ADAPFIT)
	u, err := findOrCreateUser(user, s)
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

func (s *userService) naverLogin(code string) (string, error) {
	var user model.User
	var err error
	var email string

	clientID := os.Getenv("NAVER_CLIENT_ID")
	clientSecret := os.Getenv("NAVER_CLIENT_SECRET")
	redirectURI := "http://localhost:44403/naver/callback"

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", "https://nid.naver.com/oauth2.0/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// ✅ 응답에서 ID 토큰 포함
	var tokenResponse NaverTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	// ✅ ID 토큰이 없으면 에러 반환 (설정 확인 필요)
	if tokenResponse.IDToken == "" {
		return "", fmt.Errorf("ID Token is missing. Make sure OIDC is enabled in Naver Developer Console")
	}

	// ✅ ID 토큰 검증 및 이메일 추출
	email, err = verifyNaverIDToken(tokenResponse.IDToken)
	if err != nil {
		return "", err
	}

	// ✅ 유저 데이터 설정
	user.Email = &email
	user.SnsType = uint(Naver)
	user.UserType = uint(ADAPFIT)
	u, err := findOrCreateUser(user, s)
	if err != nil {
		return "", err
	}

	// ✅ JWT 생성
	tokenString, err := generateJWT(u)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// func (service *userService) snsLogin(request LoginRequest) (string, error) {
// 	iss := decodeJwt(request.IdToken)
// 	var user model.User
// 	var err error
// 	var email string
// 	var snsType uint

// 	if strings.Contains(iss, "kakao") { // 카카오
// 		snsType = uint(Kakao)
// 		if email, err = kakaoLogin(request); err != nil {
// 			return "", err
// 		}
// 	} else if strings.Contains(iss, "google") { // 구글
// 		snsType = uint(Google)
// 		if email, err = googleLogin(request); err != nil {
// 			return "", err
// 		}
// 	} else if strings.Contains(iss, "apple") { // 애플
// 		snsType = uint(Apple)
// 		if email, err = appleLogin(request); err != nil {
// 			return "", err
// 		}
// 	}

// 	if err := copyStruct(request, &user); err != nil {
// 		return "", err
// 	}

// 	user.Email = &email
// 	user.SnsType = snsType
// 	u, err := findOrCreateUser(user, service)
// 	if err != nil {
// 		return "", err
// 	}

// 	// JWT 토큰 생성
// 	tokenString, err := generateJWT(u)
// 	if err != nil {
// 		return "", err
// 	}
// 	return tokenString, nil
// }

// func (service *userService) getUser(id uint) (UserResponse, error) {
// 	var user model.User
// 	result := service.db.Preload("Images", "type = ?", profileImageType).First(&user, id)
// 	if result.Error != nil {
// 		return UserResponse{}, errors.New("db error")
// 	}
// 	log.Println(user.CreatedAt)
// 	var userResponse UserResponse
// 	if err := copyStruct(user, &userResponse); err != nil {
// 		return UserResponse{}, err
// 	}

// 	if len(user.Images) != 0 {
// 		urlkey := extractKeyFromUrl(user.Images[0].Url, service.bucket, service.bucketUrl)
// 		thumbnailUrlkey := extractKeyFromUrl(user.Images[0].ThumbnailUrl, service.bucket, service.bucketUrl)
// 		// 사전 서명된 URL을 생성
// 		url, _ := service.s3svc.GetObjectRequest(&s3.GetObjectInput{
// 			Bucket: aws.String(service.bucket),
// 			Key:    aws.String(urlkey),
// 		})
// 		thumbnailUrl, _ := service.s3svc.GetObjectRequest(&s3.GetObjectInput{
// 			Bucket: aws.String(service.bucket),
// 			Key:    aws.String(thumbnailUrlkey),
// 		})
// 		urlStr, err := url.Presign(5 * time.Second) // URL은 5초 동안 유효
// 		if err != nil {
// 			return UserResponse{}, err
// 		}
// 		thumbnailUrlStr, err := thumbnailUrl.Presign(5 * time.Second) // URL은 5초 동안 유효 CachedNetworkImage 에서 캐싱해서 쓰면됨
// 		if err != nil {
// 			return UserResponse{}, err
// 		}
// 		userResponse.ProfileImage.Url = urlStr // 사전 서명된 URL로 업데이트
// 		userResponse.ProfileImage.ThumbnailUrl = thumbnailUrlStr

// 	}

// 	return userResponse, nil
// }

// func (service *userService) setUser(userRequest UserRequest) (string, error) {

// 	var fileName, thumbnailFileName string
// 	var image model.Image
// 	var user model.User

// 	user.ID = userRequest.ID
// 	user.Nickname = userRequest.Nickname

// 	if userRequest.ProfileImage != "" {

// 		//base64 string decode
// 		imgData, err := base64.StdEncoding.DecodeString(userRequest.ProfileImage)
// 		if err != nil {
// 			return "", err
// 		}

// 		//이미지 포맷 체크
// 		contentType, ext, err := getImageFormat(imgData)
// 		if err != nil {
// 			return "", err
// 		}

// 		// 이미지 크기 조정 (10MB 제한)
// 		if len(imgData) > 10*1024*1024 {
// 			imgData, err = reduceImageSize(imgData)
// 			if err != nil {
// 				return "", err
// 			}
// 		}

// 		// 썸네일 이미지 생성
// 		thumbnailData, err := createThumbnail(imgData)
// 		if err != nil {
// 			return "", err
// 		}

// 		// S3에 이미지 및 썸네일 업로드
// 		fileName, thumbnailFileName, err = uploadImagesToS3(imgData, thumbnailData, contentType, ext, service.s3svc, service.bucket, service.bucketUrl, strconv.FormatUint(uint64(user.ID), 10))
// 		if err != nil {

// 			return "", err
// 		}
// 		image = model.Image{
// 			Uid:          user.ID,
// 			Url:          fileName,
// 			ThumbnailUrl: thumbnailFileName,
// 			ParentId:     user.ID,
// 			Type:         uint(profileImageType),
// 		}
// 	}

// 	// 트랜잭션 시작
// 	tx := service.db.Begin()

// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			log.Printf("Recovered from panic: %v", r)
// 		}
// 	}()

// 	//유저 정보 업데이트
// 	result := tx.Model(&user).Where("id=?", user.ID).Update("nickname", user.Nickname)
// 	if result.Error != nil {
// 		log.Println(result.Error.Error())
// 		tx.Rollback()

// 		// 이미 업로드된 파일들을 S3에서 삭제

// 		if userRequest.ProfileImage != "" {
// 			go func() {
// 				deleteFromS3(fileName, service.s3svc, service.bucket, service.bucketUrl)
// 				deleteFromS3(thumbnailFileName, service.s3svc, service.bucket, service.bucketUrl)
// 			}()
// 		}

// 		return "", errors.New("db error")
// 	}
// 	if userRequest.ProfileImage != "" {
// 		// 기존 이미지 레코드 논리삭제
// 		result = tx.Where("parent_id = ? AND type =?", user.ID, profileImageType).Delete(&model.Image{})
// 		if result.Error != nil {
// 			log.Println(result.Error.Error())
// 			tx.Rollback()
// 			if userRequest.ProfileImage != "" {
// 				go func() {
// 					deleteFromS3(fileName, service.s3svc, service.bucket, service.bucketUrl)
// 					deleteFromS3(thumbnailFileName, service.s3svc, service.bucket, service.bucketUrl)
// 				}()
// 			}
// 			return "", errors.New("db error4")
// 		}
// 		// 이미지 레코드 재 생성

// 		if err := tx.Create(&image).Error; err != nil {
// 			log.Println(err)
// 			tx.Rollback()
// 			if userRequest.ProfileImage != "" {
// 				go func() {
// 					deleteFromS3(fileName, service.s3svc, service.bucket, service.bucketUrl)
// 					deleteFromS3(thumbnailFileName, service.s3svc, service.bucket, service.bucketUrl)
// 				}()
// 			}
// 			return "", errors.New("db error5")
// 		}
// 	}

// 	tx.Commit()
// 	return "200", nil
// }

// func (service *userService) removeUser(id uint) (string, error) {
// 	var user model.User
// 	user.ID = id
// 	if err := service.db.Delete(&user).Error; err != nil {
// 		return "", errors.New("db error")
// 	}
// 	return "200", nil
// }

// func (service *userService) getVersion() (AppVersionResponse, error) {
// 	var version model.AppVersion
// 	result := service.db.Last(&version)
// 	if result.Error != nil {
// 		return AppVersionResponse{}, errors.New("db error")
// 	}
// 	var versionResponse AppVersionResponse
// 	if err := copyStruct(version, &versionResponse); err != nil {
// 		return AppVersionResponse{}, err
// 	}
// 	return versionResponse, nil
// }

// func (service *userService) removeProfile(uid uint) (string, error) {

// 	// 기존 이미지 레코드 논리삭제
// 	result := service.db.Where("parent_id = ? AND type =?", uid, profileImageType).Delete(&model.Image{})
// 	if result.Error != nil {
// 		return "", errors.New("db error2")
// 	}
// 	return "200", nil
// }
