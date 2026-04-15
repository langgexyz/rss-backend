package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"rss-backend/config"
	"rss-backend/internal/fetcher"
	"rss-backend/internal/handler"
	"rss-backend/internal/model"
	"rss-backend/internal/repository"
	"rss-backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	db.AutoMigrate(&model.Feed{}, &model.Article{})

	service.InitSemaphore(cfg.Fetcher.MaxConcurrent)

	f := fetcher.New(cfg.Fetcher.TimeoutSeconds, cfg.Fetcher.MaxArticlesPerFeed)

	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)

	artSvc := service.NewArticleService(artRepo)
	ftSvc := service.NewFulltextService(artRepo)
	feedSvc := service.NewFeedService(feedRepo, artRepo, f, ftSvc)

	feedH := handler.NewFeedHandler(feedSvc, cfg.Fetcher.MinRefreshInterval)
	artH := handler.NewArticleHandler(artSvc, ftSvc)

	r := gin.Default()
	r.Use(corsMiddleware(cfg.Server.CORSOrigins))

	api := r.Group("/api")
	{
		api.POST("/feeds", feedH.Create)
		api.GET("/feeds", feedH.List)
		api.GET("/feeds/:id", feedH.GetByID)
		api.POST("/feeds/:id/refresh", feedH.Refresh)
		api.DELETE("/feeds/:id", feedH.Delete)

		api.GET("/articles", artH.List)
		api.GET("/articles/:id", artH.GetByID)
		api.GET("/articles/:id/fulltext", artH.FetchFulltext)
		api.PATCH("/articles/:id", artH.Update)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server running on %s", addr)
	r.Run(addr)
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		for _, o := range origins {
			if o == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
