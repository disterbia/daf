package main

import (
	"log"
	"os"
	"payment-service/core"
	"payment-service/model"

	_ "payment-service/docs"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"github.com/gofiber/swagger"
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

	// Redis 클라이언트 설정
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // 비밀번호가 없으면 비워둠
		DB:       0,  // Redis DB 번호
	})

	svc := core.NewPaymentService(database, redisClient)

	refundEndpoint := core.RefundEndpoint(svc)
	paymentCallbackEndpoint := core.PaymentCallbackEndpoint(svc)

	app := fiber.New()
	app.Use(logger.New())

	// Swagger 설정
	app.Get("/swagger/*", swagger.HandlerDefault) // Swagger UI 경로 설정
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// CORS 미들웨어 추가
	// app.Use(cors.New())

	app.Static("/", "./")

	app.Post("/payment/callback", core.PaymentCallbackHandler(paymentCallbackEndpoint))
	app.Post("/refund", core.RefundHandler(refundEndpoint))

	log.Fatal(app.Listen(":44403"))
}
