package core

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"
	"user-service/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

var jwtSecretKey = []byte("haruharu_mark_user")

type PublicKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []PublicKey `json:"keys"`
}

func verifyJWT(c *fiber.Ctx) (uint, string, error) {
	// 헤더에서 JWT 토큰 추출
	tokenString := c.Get("Authorization")
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

func generateJWT(user model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 1).Unix(), // 하루 유효 기간
	})

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func decodeJwt(tokenString string) string {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Println(err)
	}

	// MapClaims 타입으로 claims 확인
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// 'iss' 확인
		if iss, ok := claims["iss"].(string); ok {
			fmt.Println("Issuer (iss):", iss)
			return iss
		} else {
			fmt.Println("'iss' 이 없습니다.")
			return ""
		}
	} else {
		log.Println("클레임을 MapClaims로 변환할 수 없습니다.")
		return ""
	}
}

func validateDate(dateStr string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, errors.New("invalid date format, should be YYYY-MM-DD")
	}
	return date, nil
}

func validateUsername(username string) error {
	// 사용자명 검증 (4~20자, 특수문자 포함 불가)
	value := strings.TrimSpace(username)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9]{4,20}$`)
	if !usernameRegex.MatchString(value) {
		return errors.New("invalid username format (4~20 characters, no special characters)")
	}
	return nil
}

func validateTime(timeStr string) error {
	if len(timeStr) != 5 {
		return errors.New("invalid time format, should be HH:MM")
	}
	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		return errors.New("invalid time format, should be HH:MM")
	}
	return nil
}

func validateSignIn(request SignInRequest, snsId *string) error {

	// 전화번호 검증 (010으로 시작하는 11자리 숫자)
	phoneRegex := regexp.MustCompile(`^010\d{8}$`)
	if !phoneRegex.MatchString(request.Phone) {
		return errors.New("invalid phone format, should be 010xxxxxxxx")
	}
	if snsId != nil {
		// 아이디 검증 (4~20자, 특수문자 포함 불가)
		username := strings.TrimSpace(request.Username)
		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9]{4,20}$`)
		if !usernameRegex.MatchString(username) {
			return errors.New("invalid username format (4~20 characters, no special characters)")
		}

		// 비밀번호 검증 (최소 8~20자, 영문 대소문자/숫자/특수문자 중 2종류 이상 포함)
		if !checkPassword(request.Password) {
			return errors.New("invalid password format (must include at least two of: letters, numbers, special characters, and be at least 8 characters long)")
		}
	}

	// 사용자명 검증 (4~20자, 특수문자 포함 불가)
	name := strings.TrimSpace(request.Name)
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9]{4,20}$`)
	if !nameRegex.MatchString(name) {
		return errors.New("invalid name format (4~20 characters, no special characters)")
	}

	return nil
}

func checkPassword(value string) bool {
	password := strings.TrimSpace(value)
	// 길이 검사: 8~20자
	if len(password) < 8 || len(password) > 20 {
		return false
	}

	// 각 카테고리 포함 여부 확인
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 카테고리 중 2개 이상 포함 여부 확인
	count := 0
	if hasLower {
		count++
	}
	if hasUpper {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 2
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

func uploadImagesToS3(imgData []byte, thumbnailData []byte, contentType string, ext string, s3Client *s3.S3, bucket string, bucketUrl string, uid string) (string, string, error) {
	// 이미지 파일 이름과 썸네일 파일 이름 생성
	imgFileName := "images/profile/" + uid + "/images/" + uuid.New().String() + ext
	thumbnailFileName := "images/profile/" + uid + "/thumbnails/" + uuid.New().String() + ext

	// S3에 이미지 업로드
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(imgFileName),
		Body:        bytes.NewReader(imgData),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", "", err
	}

	// S3에 썸네일 업로드
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(thumbnailFileName),
		Body:        bytes.NewReader(thumbnailData),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", "", err
	}

	// 업로드된 이미지와 썸네일의 URL 생성 및 반환
	imgURL := "https://" + bucket + "." + bucketUrl + "/" + imgFileName
	thumbnailURL := "https://" + bucket + "." + bucketUrl + "/" + thumbnailFileName

	return imgURL, thumbnailURL, nil
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

func createThumbnail(data []byte) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// 썸네일의 크기를 절반씩 줄이면서 1MB 이하로 만듦
	for {
		newWidth := img.Bounds().Dx() / 2
		newHeight := img.Bounds().Dy() / 2

		thumbnail := resize.Resize(uint(newWidth), uint(newHeight), img, resize.Lanczos3)

		var buf bytes.Buffer
		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, thumbnail, nil)
		case "png":
			err = png.Encode(&buf, thumbnail)
		case "gif":
			err = gif.Encode(&buf, thumbnail, nil)
		case "webp":
			err = png.Encode(&buf, thumbnail)
		default:
			log.Printf("Unsupported format: %s\n", format)
			return nil, err
		}
		if err != nil {
			return nil, err
		}

		thumbnailData := buf.Bytes()
		log.Println("thumbnailData size: ", len(thumbnailData))
		if len(thumbnailData) < 1024*1024 {
			return thumbnailData, nil
		}

		img = thumbnail
	}
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
