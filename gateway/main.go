// /gateway/main.go

package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IP별 레이트 리미터를 저장할 맵과 이를 동기화하기 위한 뮤텍스
var (
	ips = make(map[string]*rate.Limiter)
	mu  sync.RWMutex
)

// 특정 IP 주소에 대한 레이트 리미터를 반환
func GetLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := ips[ip]
	if !exists {
		limiter = rate.NewLimiter(20, 20) // 레이트 리미팅 설정 조정
		ips[ip] = limiter
	}

	return limiter
}
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

// IP 주소별로 레이트 리미팅을 적용
func IPRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Swagger UI에 대한 요청은 레이트 리미팅에서 제외
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		ip := getClientIP(c)
		limiter := GetLimiter(ip)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "요청 수가 너무 많습니다",
			})
			return
		}

		c.Next()
	}
}
func main() {
	router := gin.Default()
	router.Use(IPRateLimitMiddleware())

	//서비스로의 리버스 프록시 설정
	adminServiceURL, _ := url.Parse("http://admin:44400")
	adminProxy := httputil.NewSingleHostReverseProxy(adminServiceURL)
	router.Any("/admin/*path", func(c *gin.Context) {
		c.Request.URL.Path = c.Param("path")
		adminProxy.ServeHTTP(c.Writer, c.Request)
	})

	coachServiceURL, _ := url.Parse("http://localhost:44401")
	coachProxy := httputil.NewSingleHostReverseProxy(coachServiceURL)
	router.Any("/coach/*path", func(c *gin.Context) {
		c.Request.URL.Path = c.Param("path")
		coachProxy.ServeHTTP(c.Writer, c.Request)
	})

	dafServiceURL, _ := url.Parse("http://daf:44402")
	dafProxy := httputil.NewSingleHostReverseProxy(dafServiceURL)
	router.Any("/daf/*path", func(c *gin.Context) {
		c.Request.URL.Path = c.Param("path")
		dafProxy.ServeHTTP(c.Writer, c.Request)
	})

	userServiceURL, _ := url.Parse("http://user:44403")
	userProxy := httputil.NewSingleHostReverseProxy(userServiceURL)
	router.Any("/user/*path", func(c *gin.Context) {
		c.Request.URL.Path = c.Param("path")
		userProxy.ServeHTTP(c.Writer, c.Request)
	})

	setupSwaggerUIProxy(router, "/admin-service/swagger/*proxyPath", "http://admin:44400/swagger/")
	setupSwaggerUIProxy(router, "/coach-service/swagger/*proxyPath", "http://localhost:44401/swagger/")
	setupSwaggerUIProxy(router, "/daf-service/swagger/*proxyPath", "http://daf:44402/swagger/")
	setupSwaggerUIProxy(router, "/user-service/swagger/*proxyPath", "http://user:44403/swagger/")

	// API 게이트웨이 서버 시작
	router.Run(":40000")
}

// Swagger 문서에 대한 리버스 프록시를 설정
func setupSwaggerUIProxy(router *gin.Engine, path string, target string) {
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	router.Any(path, func(c *gin.Context) {
		// Swagger 경로 재설정
		c.Request.URL.Path = c.Param("proxyPath")
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}
