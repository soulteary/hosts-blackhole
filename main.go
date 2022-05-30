//go:build go1.8
// +build go1.8

package main

import (
	"context"
	"crypto/md5"
	"embed"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/soulteary/hosts-blackhole/internal/logger"
	provider "github.com/soulteary/hosts-blackhole/provider"
)

//go:embed home.html
var EmbedHomepage embed.FS

//go:embed favicon.ico
var EmbedFavicon embed.FS

var appPort = 8080
var appDebug = false

const (
	DEFAULT_HOMEPAGE   string = "home.html"
	DEFAULT_FAVICON           = "favicon.ico"
	DEFAULT_RULES             = "rules"
	DEFAULT_LIST              = "list"
	DEFAULT_PING              = "ping"
	DEFAULT_CACHE_RULE        = "public, max-age=31536000"
	DEFAULT_CACHE_DIR         = "cache"
)

const (
	ROUTE_DEFAULT  string = "/"
	ROUTE_HOMEPAGE        = "/" + DEFAULT_HOMEPAGE
	ROUTE_FAVICON         = "/" + DEFAULT_FAVICON
	ROUTE_RULES           = "/" + DEFAULT_RULES
	ROUTE_LIST            = "/" + DEFAULT_LIST
	ROUTE_PING            = "/" + DEFAULT_PING
)

func main() {
	port := strconv.Itoa(appPort)
	if !appDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := gin.Default()
	log := logger.GetLogger()
	router.Use(logger.Logger(log), gin.Recovery())

	if !appDebug {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
		router.Use(optimizeResourceCacheTime())
	}

	router.GET(ROUTE_DEFAULT, func(c *gin.Context) {
		c.Redirect(302, ROUTE_HOMEPAGE)
		c.Abort()
	})

	router.GET(ROUTE_HOMEPAGE, func(c *gin.Context) {
		c.FileFromFS(DEFAULT_HOMEPAGE, http.FS(EmbedHomepage))
		c.Abort()
	})

	router.GET(ROUTE_FAVICON, func(c *gin.Context) {
		c.FileFromFS(DEFAULT_FAVICON, http.FS(EmbedFavicon))
		c.Abort()
	})



	router.GET(ROUTE_PING, func(c *gin.Context) {
		c.String(200, "pong")
		c.Abort()
		return
	})

	rulesDir := filepath.Join(".", "rules")
	err := os.MkdirAll(rulesDir, os.ModePerm)
	if err != nil {
		log.Fatal("程序无法创建数据目录: ", err)
	}

	cacheDir := filepath.Join(".", DEFAULT_CACHE_DIR)
	err = os.MkdirAll(cacheDir, os.ModePerm)
	if err != nil {
		log.Fatal("程序无法创建缓存目录: ", err)
	}

	router.StaticFS(ROUTE_RULES, http.Dir(DEFAULT_CACHE_DIR))

	router.GET(ROUTE_LIST, func(c *gin.Context) {

		content := provider.Test()
		provider.TriggerGC()

		err = os.WriteFile(filepath.Join(cacheDir, "hosts.txt"), []byte(strings.Join(content, "\n")), 0644)
		if err != nil {
			log.Fatal("程序无法创建缓存数据: ", err)
		}

		files, err := ioutil.ReadDir(DEFAULT_CACHE_DIR)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			fmt.Println(file.Name())
		}

		// TODO
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("design a webpage"))
		c.Abort()
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("程序启动出错: %s\n", err)
		}
	}()
	log.Println("程序已启动完毕 🚀")

	<-ctx.Done()

	stop()

	log.Println("程序正在关闭中，如需立即结束请按 CTRL+C")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("程序强制关闭: ", err)
	}

	log.Println("期待与你的再次相遇 ❤️")

}

// ViewHandler support dist handler from UI
// https://github.com/gin-gonic/gin/issues/1222
func optimizeResourceCacheTime() gin.HandlerFunc {
	data := []byte(time.Now().String())
	etag := fmt.Sprintf("W/%x", md5.Sum(data))
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, ROUTE_HOMEPAGE) ||
			strings.HasPrefix(c.Request.RequestURI, ROUTE_FAVICON) {
			c.Header("Cache-Control", DEFAULT_CACHE_RULE)
			c.Header("ETag", etag)

			if match := c.GetHeader("If-None-Match"); match != "" {
				if strings.Contains(match, etag) {
					c.Status(http.StatusNotModified)
					return
				}
			}
		}
	}
}
