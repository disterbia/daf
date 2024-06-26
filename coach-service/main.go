package main

import (
	"coach-service/core"
	"coach-service/model"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "coach-service/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	err := godotenv.Load(".env")
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

	svc := core.NewCoachService(database, s3svc, bucket, bucketUrl)
	loginEndpoint := core.LoginEndpoint(svc)
	getCategorisEndpoint := core.GetCategorisEndpoint(svc)
	saveCategoryEndpoint := core.SaveCategoryEndpoint(svc)
	getRecommendEndpoint := core.GetRecommendEndpoint(svc)
	getRecommendsEndpoint := core.GetRecommendsEndpoint(svc)
	saveExerciseEndpoint := core.SaveExerciseEndpoint(svc)
	getMachinesEndpoint := core.GetMachinesEndpoint(svc)
	saveMachineEndpoint := core.SaveMachineEndpoint(svc)
	getPurposesEndpoint := core.GetPurposesEndpoint(svc)
	saveRecommendEndpint := core.SaveRecommendEndpoint(svc)
	searchRecommendsEndpoint := core.SearchRecommendsEndpoint(svc)

	router := gin.Default()

	router.POST("/login", core.LoginHandler(loginEndpoint))
	router.GET("/get-categories", core.GetCategorisHandler(getCategorisEndpoint))
	router.POST("/save-category", core.SaveCategoryHandler(saveCategoryEndpoint))
	router.POST("/save-exercise", core.SaveExerciseHandler(saveExerciseEndpoint))
	router.GET("/get-machines", core.GetMachinesHandler(getMachinesEndpoint))
	router.POST("/save-machine", core.SaveMachineHandler(saveMachineEndpoint))
	router.GET("/get-purposes", core.GetPurposesHandler(getPurposesEndpoint))

	router.POST("/save-recommend", core.SaveRecommendHandler(saveRecommendEndpint))
	router.GET("/get-exercise/:exercise_id", core.GetRecommendHandler(getRecommendEndpoint))
	router.GET("/get-recommends", core.GetRecommendsHandler(getRecommendsEndpoint))
	router.GET("/search-recommends", core.SearchRecommendsHandler(searchRecommendsEndpoint))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":44401")
}
