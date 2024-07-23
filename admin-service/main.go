package main

import (
	"admin-service/core"
	_ "admin-service/docs"
	"admin-service/model"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ginSwagger "github.com/swaggo/gin-swagger"

	swaggerFiles "github.com/swaggo/files"
)

var ipLimiters = make(map[string]*rate.Limiter)
var ipLimitersMutex sync.Mutex

func getClientIP(c *gin.Context) string {
	// X-Real-IP 헤더를 확인
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	// X-Forwarded-For 헤더를 확인
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0] // 여러 IP가 쉼표로 구분되어 있을 수 있음
	}
	// 헤더가 없는 경우 Gin의 기본 메서드 사용
	return c.ClientIP()
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		// IP별 리미터가 있는지 확인
		ipLimitersMutex.Lock()
		limiter, exists := ipLimiters[ip]
		if !exists {
			// 새로운 리미터 생성
			limiter = rate.NewLimiter(rate.Every(time.Hour/10), 10)
			ipLimiters[ip] = limiter
		}
		ipLimitersMutex.Unlock()

		// 요청 허용 여부 확인
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "요청 횟수 초과"})
			return
		}

		c.Next()
	}
}
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

	// gRPC 클라이언트 연결 생성
	conn, err := grpc.NewClient("email:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to email service: %v", err)
	}
	defer conn.Close()

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

	svc := core.NewAdminService(database, conn, s3svc, bucket, bucketUrl)
	loginEndpoint := core.LoginEndpoint(svc)
	sendCodeEndpoint := core.SendCodeEndpoint(svc)
	verifyEndpoint := core.VerifyEndpoint(svc)
	signInEndpoint := core.SignInEndpoint(svc)
	getSupersEndpoint := core.GetSuperEndpoint(svc)
	resetEndpoint := core.ResetPasswordEndpoint(svc)
	saveEndpoint := core.SaveUserEndpoint(svc)
	searchEndpoint := core.SearhUsersEndpoint(svc)
	getAdminsEndpoint := core.GetAdminsEndpoint(svc)
	getAgencisEndpoint := core.GetAgencisEndpoint(svc)
	getDisableDetailsEndPoint := core.GetDisableDetailsEndpoint(svc)
	getAfcsEndPoint := core.GetAfcsEndpoint(svc)
	createAfcEndpoint := core.CreateAfcEndpoint(svc)
	updateAfcEndpoint := core.UpdateAfcEndpoint(svc)
	getAfcHistorisEndpoint := core.GetAfcHistorisEndpoint(svc)
	updateAfcHistoryEndpoint := core.UpdateAfcHistoryEndpoint(svc)
	searchDiaryEndpoint := core.SearhDiaryEndpoint(svc)
	saveDiaryEndPoint := core.SaveDiaryEndpoint(svc)
	getExerciseMeasuresEndpoint := core.GetExerciseMeasuresEndpoint(svc)
	getAllUsersEndpoint := core.GetAllUsersEndpoint(svc)
	getUserEndpoint := core.GetUserEndpoint(svc)
	searchMachinesEndpoint := core.SearchMachinesEndpoint(svc)
	getMachinesEndpoint := core.GetMachinesEndpoint(svc)
	saveMachinesEndpoint := core.SaveMachinesEndpoint(svc)
	removeMachinesEndpoint := core.RemoveMachinesEndpoint(svc)

	router := gin.Default()

	rateLimiterMiddleware := RateLimitMiddleware()

	router.POST("/login", core.LoginHandler(loginEndpoint))
	router.POST("/send-code/:email", rateLimiterMiddleware, core.SendCodeHandler(sendCodeEndpoint))
	router.POST("/verify-code", core.VerifyHandler(verifyEndpoint))
	router.POST("/sign-in", core.SignInHandler(signInEndpoint))
	router.POST("/reset-password", core.ResetPasswordHandler(resetEndpoint))
	router.POST("/save-user", core.SaveUserHandler(saveEndpoint))
	router.POST("/search-users", core.SearchUsersHandler(searchEndpoint))
	router.POST("/create-afc", core.CreateAfcHandler(createAfcEndpoint))
	router.POST("/update-afc", core.UpdateAfcHandler(updateAfcEndpoint))
	router.POST("/update-afc-history", core.UpdateAfcHistoryHandler(updateAfcHistoryEndpoint))
	router.POST("/search-diary", core.SearchDiaryHandler(searchDiaryEndpoint))
	router.POST("/save-diary", core.SaveDiaryHandler(saveDiaryEndPoint))
	router.POST("/search-machines", core.SearchMachinesHandler(searchMachinesEndpoint))
	router.POST("/save-machines", core.SaveMachinesHandler(saveMachinesEndpoint))
	router.POST("/remove-machines", core.RemoveDiaryHandler(removeMachinesEndpoint))

	router.GET("/get-supers", core.GetSupersHandler(getSupersEndpoint))
	router.GET("/get-admins", core.GetAdminsHandler(getAdminsEndpoint))
	router.GET("/get-agencis", core.GetAgencisHandler(getAgencisEndpoint))
	router.GET("/get-details", core.GetDisableDetailsHandler(getDisableDetailsEndPoint))
	router.GET("/get-afcs/:id", core.GetAfcsHandler(getAfcsEndPoint))
	router.GET("/get-user/:id", core.GetUserHandler(getUserEndpoint))
	router.GET("/get-historis/:id", core.GetAfcHistorisHandler(getAfcHistorisEndpoint))
	router.GET("/get-all-users", core.GetAllUsersHandler(getAllUsersEndpoint))
	router.GET("/get-exercise-measures", core.GetExerciseMeasuresHandler(getExerciseMeasuresEndpoint))
	router.GET("/get-machines", core.GetMachinesHandler(getMachinesEndpoint))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":44400")
}
