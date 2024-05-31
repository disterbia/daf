package main

import (
	"coach-service/core"
	"coach-service/model"
	"log"
	"os"

	"github.com/gin-contrib/cors"
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

	svc := core.NewCoachService(database)
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

	router := gin.Default()

	config := cors.Config{
		AllowAllOrigins:  true, // 모든 출처 허용
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	router.Use(cors.New(config))

	router.POST("/login", core.LoginHandler(loginEndpoint))
	router.GET("/get-categories", core.GetCategorisHandler(getCategorisEndpoint))
	router.POST("/save-category", core.SaveCategoryHandler(saveCategoryEndpoint))
	router.POST("/save-exercise", core.SaveExerciseHandler(saveExerciseEndpoint))
	router.GET("/get-machines", core.GetMachinesHandler(getMachinesEndpoint))
	router.POST("/save-machine", core.SaveMachineHandler(saveMachineEndpoint))
	router.GET("/get-purposes", core.GetPurposesHandler(getPurposesEndpoint))

	router.POST("/save-recommend", core.SaveRecommendHandler(saveRecommendEndpint))
	router.GET("/get-recommend/:exercise_id", core.GetRecommendHandler(getRecommendEndpoint))
	router.GET("/get-recommends", core.GetRecommendsHandler(getRecommendsEndpoint))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":44401")
}
