package core

import (
	"bytes"
	"coach-service/model"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

var jwtSecretKey = []byte("adapfit_mark")

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

func copyStruct(input interface{}, output interface{}) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, output)
	if err != nil {
		return err
	}

	return nil
}

func validateRecommendRequest(request RecommendRequest) error {
	if request.ExerciseID == 0 || request.MachineIDs == nil || request.PurposeIDs == nil || len(request.MachineIDs) == 0 || len(request.PurposeIDs) == 0 ||
		request.BodyType > 3 || request.BodyType == 0 || request.TrRom == 0 || request.Locomotion == 0 || request.Afcs == nil || len(request.Afcs) == 0 {
		return errors.New("check body")
	}

	if request.BodyType == uint(TBODY) && len(request.Afcs) != 4 {
		return errors.New("afcs length must be 4")
	}

	if (request.BodyType == uint(UBODY) || request.BodyType == uint(LBODY)) && len(request.Afcs) != 2 {
		return errors.New("afcs length must be 2")
	}

	var checkJoint = make(map[uint]bool)

	for _, v := range request.Afcs {
		if request.BodyType == uint(UBODY) && (v.JointAction == uint(HIP) || v.JointAction == uint(KNEE)) {
			return errors.New("check body0")
		}
		if request.BodyType == uint(LBODY) && (v.JointAction == uint(SHOULDER) || v.JointAction == uint(ELBOW)) {
			return errors.New("check body1")
		}
		if v.JointAction > 4 || v.JointAction == 0 {
			return errors.New("check body2")
		}

		if v.Rom == 0 || v.ClinicDegree == nil || len(v.ClinicDegree) != len(CLINIC) {
			return errors.New("check body3")
		}

		var checkClinic = make(map[uint]bool)
		checkAC := true
		for clinic, degree := range v.ClinicDegree {
			if degree != 1 {
				checkAC = false
			}
			if _, exists := checkClinic[clinic]; exists {
				return errors.New("duplicate clinic") // 중복된 clinic인 경우 처리하지 않음
			}
			checkClinic[clinic] = true
			if clinic == uint(AC) && !((degree == 1 && v.Rom == 1) || (degree == 5 && v.Rom == 5)) {
				return errors.New("must 1 or 5") // 절단은 반드시 1 또는 5
			}
			if clinic == uint(AC) && degree == 1 && v.Rom == 1 {
				if !checkAC { //ac일땐 반드시 모두 1
					return errors.New("must 1")
				}
			}
		}

		if _, exists := checkJoint[v.JointAction]; exists {
			return errors.New("duplicate joint") // 중복된 jointaction 경우 처리하지 않음
		}
		checkJoint[v.JointAction] = true
	}

	return nil
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
