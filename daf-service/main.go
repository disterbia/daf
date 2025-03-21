package main

import (
	"daf-service/core"
	_ "daf-service/docs"
	"daf-service/model"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	ginSwagger "github.com/swaggo/gin-swagger"

	swaggerFiles "github.com/swaggo/files"
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

	svc := core.NewDafService(database)
	getRecommendEndpoint := core.GetRecommendsEndpoint(svc)
	fmt.Printf("svc: %v\n", svc)

	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/get-recommend", core.GetRecommendsHandler(getRecommendEndpoint))

	router.Run(":44402")
}
