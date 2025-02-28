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
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	// lis, err := net.Listen("tcp", ":50052")
	// if err != nil {
	// 	log.Fatalf("failed to listen: %v", err)
	// }

	// grpcServer := grpc.NewServer()
	// pb.RegisterUserServiceServer(grpcServer, &core.UserServer{DB: database, S3svc: s3svc, Bucket: bucket, BucketUrl: bucketUrl})

	// if err := grpcServer.Serve(lis); err != nil {
	// 	log.Fatalf("failed to serve: %v", err)
	// }

	svc := core.NewUserService(database, s3svc, bucket, bucketUrl)

	// snsLoginEndpoint := core.SnsLoginEndpoint(svc)
	// getUserEndpoint := core.GetUserEndpoint(svc)
	// setUserEndpoint := core.SetUserEndpoint(svc)
	// removeEndpoint := core.RemoveEndpoint(svc)
	// removeProfileEndpoint := core.RemoveProfileEndpoint(svc)

	appleCallbackEndpoint := core.AppleCallbackEndpoint(svc)
	googleCallbackEndpoint := core.GoogleCallbackEndpoint(svc)
	kakaoCallbackEndpoint := core.KakaoCallbackEndpoint(svc)
	facebookCallbackEndpoint := core.FacebookCallbackEndpoint(svc)
	naverCallbackEndpoint := core.NaverCallbackEndpoint(svc)
	app := fiber.New()
	app.Use(logger.New())

	// Swagger 설정
	app.Get("/swagger/*", swagger.HandlerDefault) // Swagger UI 경로 설정
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// CORS 미들웨어 추가
	app.Use(cors.New())

	// app.Get("/get-user", core.GetUserHandler(getUserEndpoint))

	// app.Post("/sns-login", core.SnsLoginHandler(snsLoginEndpoint))
	// app.Post("/set-user", core.SetUserHandler(setUserEndpoint))
	// app.Post("/remove-user", core.RemoveHandler(removeEndpoint))
	// app.Post("/remove-profile", core.RemoveProfileHandler(removeProfileEndpoint))

	app.Post("/apple/callback", core.AppleCallbackHandler(appleCallbackEndpoint))
	app.Get("/google/callback", core.GoogleCallbackHandler(googleCallbackEndpoint))
	app.Get("/kakao/callback", core.KakaoCallbackHandler(kakaoCallbackEndpoint))
	app.Get("/facebook/callback", core.FacebookCallbackHandler(facebookCallbackEndpoint))
	app.Get("/naver/callback", core.NaverCallbackHandler(naverCallbackEndpoint))

	log.Fatal(app.Listen(":44403"))
}
