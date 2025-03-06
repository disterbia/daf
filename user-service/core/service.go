// /user-service/service/service.go

package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"user-service/model"
	pb "user-service/proto"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	appleLogin(code string) (LoginResponse, error)
	googleLogin(code string) (LoginResponse, error)
	kakaoLogin(code string) (LoginResponse, error)
	facebookLogin(code string) (LoginResponse, error)
	naverLogin(code string) (LoginResponse, error)

	checkUsername(username string) (string, error)
	basicLogin(request LoginRequest) (string, error)
	signIn(request SignInRequest) (string, error)
	sendAuthCode(phone string) (string, error)
	verifyAuthCode(phone string, code string) (string, error)
	getUser(id uint) (UserResponse, error) //유저조회
	findUsername(request FindUsernameRequest) (string, error)
	findPassword(request FindPasswordRequest) (string, error)
	setUser(request SetUserRequest) (string, error)

	// snsLogin(request LoginRequest) (string, error)

	// setUser(userRequest UserRequest) (string, error)
	// removeUser(id uint) (string, error)
	// getVersion() (AppVersionResponse, error)
	// removeProfile(uid uint) (string, error)
}

type userService struct {
	db          *gorm.DB
	s3svc       *s3.S3
	bucket      string
	bucketUrl   string
	redisClient *redis.Client
}

func NewUserService(db *gorm.DB, s3svc *s3.S3, bucket string, bucketUrl string, redisClient *redis.Client) UserService {
	return &userService{db: db, s3svc: s3svc, bucket: bucket, bucketUrl: bucketUrl, redisClient: redisClient}
}

type UserServer struct {
	pb.UnimplementedUserServiceServer
	DB        *gorm.DB
	S3svc     *s3.S3
	Bucket    string
	BucketUrl string
}

func (s *userService) checkUsername(username string) (string, error) {
	var user model.User
	if err := validateUsername(username); err != nil {
		return "", err
	}
	if err := s.db.Where("username = ?", username).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return "-1", nil
	} else if err != nil {
		return "", errors.New("db error")
	}
	return "1", nil

}
func (service *userService) sendAuthCode(phone string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 함수 종료 시 컨텍스트 해제
	// 6자리 랜덤 인증번호 생성
	authCode := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	// Redis에 인증번호 저장 (유효시간: 5분)
	err := service.redisClient.Set(ctx, phone, authCode, time.Minute*5).Err()
	if err != nil {
		log.Printf("Failed to save auth code in Redis: %v", err)
		return "", errors.New("failed to send auth code")
	}

	if err := sendCode(phone, authCode); err != nil {
		return "", err
	}

	log.Printf("Auth code for %s is %s", phone, authCode)

	return "1", nil
}

func (service *userService) verifyAuthCode(phone string, code string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 함수 종료 시 컨텍스트 해제

	// Redis에서 인증번호 조회
	storedCode, err := service.redisClient.Get(ctx, phone).Result()
	if err == redis.Nil {
		return "-1", nil
	} else if err != nil {
		log.Printf("Failed to get auth code: %v", err)
		return "", errors.New("internal error")
	}

	// 입력된 코드와 비교
	if storedCode == code {
		// 인증 성공 시 Redis에 "인증 완료" 상태 플래그 설정
		err := service.redisClient.Set(ctx, phone+":status", "verified", time.Minute*10).Err()
		if err != nil {
			return "", errors.New("failed to set verified status")
		}
		// 기존 인증번호는 삭제하지 않고 그대로 둠
		return "1", nil
	}

	return "-1", nil
}

func (s *userService) signIn(request SignInRequest) (string, error) {
	var gender uint
	var snsType uint
	var isAgree uint

	var snsId *string
	var password *string
	var username *string

	if request.Gender {
		gender = 1
	} else {
		gender = 2
	}
	if request.IsAgree {
		isAgree = 1
	} else {
		isAgree = 2
	}

	birthday, err := time.Parse("2006-01-02", request.Birth)
	if err != nil {
		return "", errors.New("date err")
	}

	// 단일 컨텍스트 생성 (타임아웃 5초 설정)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 함수 종료 시 컨텍스트 해제

	// Redis에서 키 검색
	snsTypeValue, err := s.redisClient.Get(ctx, request.SnsId).Result()
	if err == redis.Nil {
		snsType = uint(Password)
	} else if err != nil {
		return "", errors.New("internal error")
	} else {
		value, err := strconv.ParseUint(snsTypeValue, 10, 32)
		if err != nil {
			return "", errors.New("invalid value")
		}
		snsId = &request.SnsId
		snsType = uint(value)
	}

	if err := validateSignIn(request, snsId); err != nil {
		return "", err
	}

	if snsId != nil {
		username = &request.Username
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		temp := string(hashedPassword)
		password = &temp
	}

	phoneStatusKey := request.Phone + ":status"
	// Redis에서 인증 상태 확인
	if status, err := s.redisClient.Get(ctx, phoneStatusKey).Result(); err == redis.Nil || status != "verified" {
		return "-1", nil
	} else if err != nil {
		log.Printf("Failed to check verification status: %v", err)
		return "", errors.New("internal error2")
	}

	// Redis에 존재하면 user 테이블에 추가
	var user = model.User{
		SnsId:      snsId,
		Username:   username,
		Password:   password,
		Name:       request.Name,
		Birthday:   birthday,
		Phone:      request.Phone,
		Gender:     gender,
		IsAgree:    isAgree,
		Addr:       request.Addr,
		AddrDetail: request.AddrDetail,
		UserType:   uint(ADAPFIT),
		SnsType:    uint(snsType), // redis value
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		} else if tx.Error != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&user).Error; err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "phone") {
			var existUser model.User
			if err := tx.Where("phone = ?", user.Phone).First(&existUser).Error; err != nil {
				return "", errors.New("db error")
			}
			user.ID = existUser.ID
			if err := tx.Model(&existUser).
				Select("SnsId", "Username", "Password", "Name", "Birthday", "Gender", "Addr", "AddrDetail", "UserType", "SnsType").
				Updates(user).Error; err != nil {
				return "", errors.New("db error2")
			}
		} else {
			tx.Rollback()
			return "", errors.New("db error3")
		}
	}

	if err := tx.Create(&model.UserVisit{UID: user.ID, VisitPurposeID: request.VisitPurpose}).Error; err != nil {
		return "", errors.New("db error4")
	}
	if err := tx.Create(&model.UserDisable{UID: user.ID, DisableTypeID: request.DisableType}).Error; err != nil {
		return "", errors.New("db error5")
	}

	if err = s.redisClient.Del(ctx, request.SnsId, phoneStatusKey).Err(); err != nil {
		tx.Rollback()
		log.Println(err)
		return "", errors.New("internal error3")
	}

	tx.Commit()
	return "1", nil
}

func (s *userService) basicLogin(request LoginRequest) (string, error) {
	var u model.User
	password := strings.TrimSpace(request.Password)

	if password == "" {
		return "", errors.New("empty")
	}

	// 이메일로 사용자 조회
	if err := s.db.Where("username = ?", request.Username).First(&u).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return "-1", nil
	} else if err != nil {
		return "", errors.New("db error")
	}

	// 비밀번호 비교
	if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(request.Password)); err != nil {
		return "-1", nil
	}

	// 새로운 JWT 토큰 생성
	tokenString, err := generateJWT(u)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) findUsername(request FindUsernameRequest) (string, error) {
	// 단일 컨텍스트 생성 (타임아웃 5초 설정)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 함수 종료 시 컨텍스트 해제

	phoneStatusKey := request.Phone + ":status"
	// Redis에서 인증 상태 확인
	if status, err := s.redisClient.Get(ctx, phoneStatusKey).Result(); err == redis.Nil || status != "verified" {
		return "-1", nil
	} else if err != nil {
		log.Printf("Failed to check verification status: %v", err)
		return "", errors.New("internal error")
	}

	var user model.User
	if err := s.db.Where("name = ? AND phone = ?", request.Name, request.Phone).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return "-2", nil
	} else if err != nil {
		return "", errors.New("db error")
	}

	if err := s.redisClient.Del(ctx, phoneStatusKey).Err(); err != nil {
		log.Println(err)
		return "", errors.New("internal error2")
	}

	if user.Username == nil {
		return "-2", nil
	}

	return *user.Username, nil
}

func (s *userService) findPassword(request FindPasswordRequest) (string, error) {
	// 단일 컨텍스트 생성 (타임아웃 5초 설정)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 함수 종료 시 컨텍스트 해제

	phoneStatusKey := request.Phone + ":status"
	// Redis에서 인증 상태 확인
	if status, err := s.redisClient.Get(ctx, phoneStatusKey).Result(); err == redis.Nil || status != "verified" {
		return "-1", nil
	} else if err != nil {
		log.Printf("Failed to check verification status: %v", err)
		return "", errors.New("internal error")
	}

	var user model.User
	if err := s.db.Where("username = ? AND phone = ?", request.Username, request.Phone).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return "-2", nil
	} else if err != nil {
		return "", errors.New("db error")
	}

	if err := s.redisClient.Del(ctx, phoneStatusKey).Err(); err != nil {
		log.Println(err)
		return "", errors.New("internal error2")
	}

	// 새로운 JWT 토큰 생성
	tokenString, err := generateJWT(user)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) setUser(request SetUserRequest) (string, error) {
	var password *string
	if request.Password != "" {
		if !checkPassword(request.Password) {
			return "", errors.New("invalid password format (must include at least two of: letters, numbers, special characters, and be at least 8 characters long)")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		temp := string(hashedPassword)
		password = &temp

		if err := s.db.Where("id = ? ", request.Uid).Update("password", password).Error; err != nil {
			return "", errors.New("db error")
		}
		return "1", nil
	}

	var user model.User
	if err := s.db.Where("id = ?", request.Uid).First(&user).Error; err != nil {
		return "", errors.New("db error1")
	}
	if user.Phone == request.Phone {
		var newUser = model.User{Name: request.Name, Addr: request.Addr, AddrDetail: request.AddrDetail}
		if err := s.db.Model(&user).
			Select("Name", "Addr", "AddrDetail").
			Updates(newUser).Error; err != nil {
			return "", errors.New("db error2")
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel() // 함수 종료 시 컨텍스트 해제
		phoneStatusKey := request.Phone + ":status"
		// Redis에서 인증 상태 확인
		if status, err := s.redisClient.Get(ctx, phoneStatusKey).Result(); err == redis.Nil || status != "verified" {
			return "-1", nil
		} else if err != nil {
			log.Printf("Failed to check verification status: %v", err)
			return "", errors.New("internal error")
		}
		var newUser = model.User{Name: request.Name, Phone: request.Phone, Addr: request.Addr, AddrDetail: request.AddrDetail}
		if err := s.db.Model(&user).
			Select("Name", "Addr", "Phone", "AddrDetail").
			Updates(newUser).Error; err != nil {
			return "", errors.New("db error3")
		}

		if err := s.redisClient.Del(ctx, phoneStatusKey).Err(); err != nil {
			log.Println(err)
			return "", errors.New("internal error2")
		}
	}
	return "1", nil

}

func (s *userService) getUser(id uint) (UserResponse, error) {
	var user model.User
	if err := s.db.Where("id = ? ", id).Preload("UserDisables").Preload("UserVisits").First(&user).Error; err != nil {
		return UserResponse{}, err
	}
	gender := true
	if user.Gender == 2 {
		gender = false
	}
	isAgree := true
	if user.IsAgree == 2 {
		isAgree = false
	}
	response := UserResponse{Username: *user.Username, Name: user.Name, Gender: gender, Birth: user.Birthday.Format("2006-01-02"), IsAgree: isAgree,
		Phone: user.Phone, Addr: user.Addr, AddrDetail: user.AddrDetail, DisableType: user.UserDisables[0].DisableTypeID, VisitPurpose: user.UserVisits[0].VisitPurposeID}
	return response, nil
}

func (s *userService) appleLogin(code string) (LoginResponse, error) {
	var err error
	var snsId string

	clientID := os.Getenv("APPLE_CLIENT_ID")
	keyID := os.Getenv("APPLE_KEY_ID")
	teamID := os.Getenv("APPLE_TEAM_ID")
	privateKey := os.Getenv("APPLE_PRIVATE_KEY")

	// client_secret 생성
	clientSecret, err := GenerateClientSecret(keyID, teamID, clientID, privateKey)
	if err != nil {
		return LoginResponse{}, err
	}

	// 요청 데이터 생성
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", "https://localhost:44403/apple/callback")

	// POST 요청 생성
	req, err := http.NewRequest("POST", "https://appleid.apple.com/auth/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return LoginResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return LoginResponse{}, err
	}
	defer resp.Body.Close()

	// 응답 확인
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return LoginResponse{}, fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 응답 파싱
	var tokenResponse AppleTokenResponse
	tokenResponse.Code = code
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return LoginResponse{}, err
	}
	if snsId, err = appleLogin(tokenResponse.IDToken); err != nil {
		return LoginResponse{}, err
	}

	response, err := snsLogin(snsId, uint(Apple), s)
	if err != nil {
		return LoginResponse{}, err
	}
	return response, nil
}

func (s *userService) googleLogin(code string) (LoginResponse, error) {
	var err error
	var snsId string

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := "http://localserver.com:44403/google/callback"

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
		return LoginResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return LoginResponse{}, err
	}
	defer resp.Body.Close()

	// 응답 확인
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return LoginResponse{}, fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 응답 파싱
	var tokenResponse GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return LoginResponse{}, err
	}

	// ID 토큰 검증 및 sub 추출
	if snsId, err = googleLogin(tokenResponse.IDToken, clientID); err != nil {
		return LoginResponse{}, err
	}

	response, err := snsLogin(snsId, uint(Apple), s)
	if err != nil {
		return LoginResponse{}, err
	}

	return response, nil
}

func (s *userService) kakaoLogin(code string) (LoginResponse, error) {
	var err error
	var snsId string

	clientID := os.Getenv("KAKAO_CLIENT_ID")
	clientSecret := os.Getenv("KAKAO_CLIENT_SECRET")
	redirectURI := "http://192.168.0.18:44403/kakao/callback"

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
		return LoginResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 요청 전송
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return LoginResponse{}, err
	}
	defer resp.Body.Close()

	// 응답 확인
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return LoginResponse{}, fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 응답 파싱
	var tokenResponse KakaoTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return LoginResponse{}, err
	}

	// ID 토큰 검증 및 이메일 추출
	if snsId, err = kakaoLogin(tokenResponse.IDToken); err != nil {
		return LoginResponse{}, err
	}

	response, err := snsLogin(snsId, uint(Apple), s)
	if err != nil {
		return LoginResponse{}, err
	}

	return response, nil
}

func (s *userService) facebookLogin(code string) (LoginResponse, error) {
	var err error
	var snsId string

	clientID := os.Getenv("FACEBOOK_CLIENT_ID")
	clientSecret := os.Getenv("FACEBOOK_CLIENT_SECRET")
	redirectURI := "http://192.168.0.18:44403/facebook/callback"

	// Access Token 요청
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("code", code)

	req, err := http.NewRequest("POST", "https://graph.facebook.com/v18.0/oauth/access_token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return LoginResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return LoginResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return LoginResponse{}, fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 응답 JSON을 구조체로 변환
	var tokenResponse FacebookTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return LoginResponse{}, err
	}

	// 액세스 토큰이 없으면 에러 반환
	if tokenResponse.AccessToken == "" {
		return LoginResponse{}, fmt.Errorf("failed to retrieve access token from Facebook")
	}

	// 페이스북 사용자 정보 요청
	snsId, err = getFacebookUserInfo(tokenResponse.AccessToken)
	if err != nil {
		return LoginResponse{}, err
	}

	// 유저 데이터 처리
	response, err := snsLogin(snsId, uint(Apple), s)
	if err != nil {
		return LoginResponse{}, err
	}

	return response, nil
}

func (s *userService) naverLogin(code string) (LoginResponse, error) {
	var err error
	var snsId string

	clientID := os.Getenv("NAVER_CLIENT_ID")
	clientSecret := os.Getenv("NAVER_CLIENT_SECRET")
	redirectURI := "http://192.168.0.18:44403/naver/callback"

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", "https://nid.naver.com/oauth2.0/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return LoginResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return LoginResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return LoginResponse{}, fmt.Errorf("failed to exchange token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// ✅ 응답에서 ID 토큰 포함
	var tokenResponse NaverTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return LoginResponse{}, err
	}

	if tokenResponse.AccessToken == "" {
		return LoginResponse{}, fmt.Errorf("AccessToken miss")
	}
	snsId, err = getNaverUserInfo(tokenResponse.AccessToken)
	if err != nil {
		return LoginResponse{}, err
	}

	// 유저 데이터 처리
	response, err := snsLogin(snsId, uint(Apple), s)
	if err != nil {
		return LoginResponse{}, err
	}

	return response, nil
}
