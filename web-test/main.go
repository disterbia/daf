package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 프리렌더링 미들웨어 추가
	r.Use(prerenderMiddleware)

	// 정적 파일 제공
	r.Static("/assets", "./build/web/assets")
	r.Static("/canvaskit", "./build/web/canvaskit")
	r.StaticFile("/main.dart.js", "./build/web/main.dart.js")
	r.StaticFile("/flutter_service_worker.js", "./build/web/flutter_service_worker.js")
	r.StaticFile("/flutter_bootstrap.js", "./build/web/flutter_bootstrap.js")
	r.StaticFile("/manifest.json", "./build/web/manifest.json")
	r.StaticFile("/favicon.png", "./build/web/favicon.png")
	r.StaticFile("/icons/Icon-192.png", "./build/web/icons/Icon-192.png")

	// Index.html을 제공
	r.LoadHTMLFiles("./build/web/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 모든 경로에 대해 프리렌더링 적용
	r.NoRoute(func(c *gin.Context) {
		if isBot(c.Request.UserAgent()) {
			ctx, cancel := chromedp.NewContext(context.Background())
			defer cancel()

			var res string
			url := fmt.Sprintf("http://localhost:40000%s", c.Request.RequestURI)
			err := chromedp.Run(ctx,
				chromedp.Navigate(url),
				chromedp.OuterHTML("html", &res),
			)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error rendering page")
				return
			}

			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(res))
			c.Abort()
		} else {
			c.HTML(http.StatusOK, "index.html", nil)
		}
	})

	r.Run(":40000")
}

func prerenderMiddleware(c *gin.Context) {
	if isBot(c.Request.UserAgent()) {
		ctx, cancel := chromedp.NewContext(context.Background())
		defer cancel()

		var res string
		url := fmt.Sprintf("http://localhost:40000%s", c.Request.RequestURI)
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.OuterHTML("html", &res),
		)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error rendering page")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(res))
		c.Abort()
	} else {
		c.Next()
	}
}

func isBot(userAgent string) bool {
	bots := []string{"Googlebot", "Bingbot", "Yahoo"}
	for _, bot := range bots {
		if strings.Contains(userAgent, bot) {
			return true
		}
	}
	return false
}
