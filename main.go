package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirius/bt_search/api"
	"github.com/sirius/bt_search/config"
	"github.com/sirius/bt_search/models"
	"github.com/sirius/bt_search/utils"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化Elasticsearch客户端
	if err := utils.InitElasticsearch(); err != nil {
		log.Fatalf("初始化Elasticsearch失败: %v", err)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 注册自定义模板函数
	r.SetFuncMap(template.FuncMap{
		"formatSize": formatSize,
		"formatDate": formatDate,
		"add":        func(a, b int) int { return a + b },
	})

	// 加载HTML模板
	r.LoadHTMLGlob("templates/*")

	// 静态文件服务
	r.Static("/static", "./static")

	// API路由
	r.GET("/api/search", api.SearchTorrents)
	r.GET("/api/torrent/:infohash", api.GetTorrent)

	// 前端页面路由
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/torrent/:infohash", func(c *gin.Context) {
		infoHash := c.Param("infohash")
		if infoHash == "" {
			c.HTML(http.StatusNotFound, "index.html", gin.H{
				"error": "未找到该种子",
			})
			return
		}

		// 记录日志，帮助调试
		log.Printf("访问种子详情页，InfoHash: %s", infoHash)

		// 构建查询以获取指定info_hash的种子
		req := models.SearchRequest{
			Query: "info_hash:" + infoHash,
			Page:  1,
		}

		response, err := utils.SearchTorrents(req)
		if err != nil {
			log.Printf("搜索错误: %v", err)
			c.HTML(http.StatusNotFound, "index.html", gin.H{
				"error": "搜索出错: " + err.Error(),
			})
			return
		}

		if len(response.Results) == 0 {
			log.Printf("未找到种子，InfoHash: %s", infoHash)
			c.HTML(http.StatusNotFound, "index.html", gin.H{
				"error": "未找到该种子",
			})
			return
		}

		log.Printf("成功获取种子详情，名称: %s", response.Results[0].Name)
		c.HTML(http.StatusOK, "detail.html", response.Results[0])
	})

	// 启动服务器
	port := config.AppConfig.Server.Port
	log.Printf("服务器启动在 http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// formatDate 格式化日期
func formatDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
