package main

import (
	"log"
	"os"
	"time"
	"user_service/core"
	"user_service/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	ginSwagger "github.com/swaggo/gin-swagger"

	swaggerFiles "github.com/swaggo/files"
)

func main() {
	err := godotenv.Load(".env")
	currentTime := time.Now()
	log.Println(currentTime)
	if err != nil {
		log.Println("Error loading .env file")
	}
	dbPath := os.Getenv("DB_PATH")
	database, err := model.NewDB(dbPath)
	if err != nil {
		log.Println("Database connection error:", err)
	}

	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	bucketUrl := os.Getenv("S3_BUCKET_URL")
	s3sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-northeast-2"),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Println("aws connection error:", err)
	}
	s3svc := s3.New(s3sess)

	usvc := core.NewUserService(database, s3svc, bucket, bucketUrl)

	snsLoginEndpoint := core.SnsLoginEndpoint(usvc)
	autoLoginEndpoint := core.AutoLoginEndpoint(usvc)
	getUserEndpoint := core.GetUserEndpoint(usvc)
	setUserEndpoint := core.SetUserEndpoint(usvc)
	removeEndpoint := core.RemoveEndpoint(usvc)
	removeProfileEndpoint := core.RemoveProfileEndpoint(usvc)
	getversionEndpoint := core.GetVersionEndpoint(usvc)

	router := gin.Default()

	router.POST("/sns-login", core.SnsLoginHandler(snsLoginEndpoint))
	router.POST("/auto-login", core.AutoLoginHandler(autoLoginEndpoint))
	router.POST("/set-user", core.SetUserHandler(setUserEndpoint))
	router.GET("/get-user", core.GetUserHandler(getUserEndpoint))
	router.POST("/remove-user", core.RemoveHandler(removeEndpoint))
	router.POST("/remove-profile", core.RemoveProfileHandler(removeProfileEndpoint))
	router.GET("/get-version", core.GetVersionHandeler(getversionEndpoint))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":44409")
}
