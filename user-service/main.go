package main

import (
	"log"
	"os"
	"user-service/core"
	"user-service/model"

	_ "user-service/docs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

	// Redis 클라이언트 설정
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // 비밀번호가 없으면 비워둠
		DB:       0,  // Redis DB 번호
	})

	svc := core.NewUserService(database, s3svc, bucket, bucketUrl, redisClient)

	appleCallbackEndpoint := core.AppleCallbackEndpoint(svc)
	googleCallbackEndpoint := core.GoogleCallbackEndpoint(svc)
	kakaoCallbackEndpoint := core.KakaoCallbackEndpoint(svc)
	facebookCallbackEndpoint := core.FacebookCallbackEndpoint(svc)
	naverCallbackEndpoint := core.NaverCallbackEndpoint(svc)

	checkUsernameEndpoint := core.CheckUsernameEndpoint(svc)
	basicLogionEndpoint := core.BaiscLoginEndpoint(svc)
	signInEndpoint := core.SignInEndpoint(svc)
	sendCodeEndpoint := core.SendCodeEndpoint(svc)
	verifyEndpoint := core.VerifyEndpoint(svc)
	getUserEndpoint := core.GetUserEndpoint(svc)
	findUsernameEndpoint := core.FinUsernameEndpoint(svc)
	findPasswordEndpoint := core.FindPasswordEndpoint(svc)
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

	app.Get("/get-user", core.GetUserHandler(getUserEndpoint))
	app.Get("/google/callback", core.GoogleCallbackHandler(googleCallbackEndpoint))
	app.Get("/kakao/callback", core.KakaoCallbackHandler(kakaoCallbackEndpoint))
	app.Get("/facebook/callback", core.FacebookCallbackHandler(facebookCallbackEndpoint))
	app.Get("/naver/callback", core.NaverCallbackHandler(naverCallbackEndpoint))

	app.Post("/apple/callback", core.AppleCallbackHandler(appleCallbackEndpoint))
	app.Get("/check-username", core.CheckUsernameHandler(checkUsernameEndpoint))
	app.Post("/login", core.BasicLoginHandler(basicLogionEndpoint))
	app.Post("/sign-in", core.SignInHandler(signInEndpoint))
	app.Post("/send-code", core.SendCodeHandler(sendCodeEndpoint))
	app.Post("/verify-code", core.VerifyHandler(verifyEndpoint))
	app.Post("/find-username", core.FindUsernameHandler(findUsernameEndpoint))
	app.Post("/find-password", core.FindPasswordHandler(findPasswordEndpoint))
	app.Post("/payment/callback", core.PaymentCallbackHandler(paymentCallbackEndpoint))
	app.Post("/refund", core.RefundHandler(refundEndpoint))

	log.Fatal(app.Listen(":44403"))
}
