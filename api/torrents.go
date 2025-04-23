package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirius/bt_search/models"
	"github.com/sirius/bt_search/utils"
)

// SearchTorrents 处理种子搜索请求
func SearchTorrents(c *gin.Context) {
	var req models.SearchRequest

	// 设置默认值
	req.Page = 1
	req.SortBy = ""
	req.SortOrder = "desc"

	// 解析查询参数
	req.Query = c.Query("q")
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}
	req.SortBy = c.Query("sort")
	if order := c.Query("order"); order != "" {
		req.SortOrder = order
	}

	// 执行搜索
	response, err := utils.SearchTorrents(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTorrent 获取单个种子的详细信息
func GetTorrent(c *gin.Context) {
	infoHash := c.Param("infohash")
	if infoHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的 info_hash",
		})
		return
	}

	// 创建一个只查询指定 info_hash 的请求
	req := models.SearchRequest{
		Query: "info_hash:" + infoHash,
		Page:  1,
	}

	response, err := utils.SearchTorrents(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取种子信息失败: " + err.Error(),
		})
		return
	}

	if len(response.Results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "未找到该种子",
		})
		return
	}

	c.JSON(http.StatusOK, response.Results[0])
}
