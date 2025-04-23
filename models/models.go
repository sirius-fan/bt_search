package models

import (
	"time"
)

// TorrentFile 表示种子中的单个文件
type TorrentFile struct {
	Path   string `json:"path"`
	Length int64  `json:"length"`
}

// TorrentMetadata 表示一个种子的元数据
type TorrentMetadata struct {
	InfoHash   string        `json:"info_hash"`
	Name       string        `json:"name"`
	Files      []TorrentFile `json:"files"`
	TotalSize  int64         `json:"total_size"`
	FileCount  int           `json:"file_count"`
	CreateDate time.Time     `json:"create_date"`
}

// SearchRequest 表示搜索请求参数
type SearchRequest struct {
	Query     string `form:"q"`
	Page      int    `form:"page" binding:"min=1"`
	SortBy    string `form:"sort" binding:"omitempty,oneof=date size files"`
	SortOrder string `form:"order" binding:"omitempty,oneof=asc desc"`
}

// SearchResponse 表示搜索响应
type SearchResponse struct {
	Total   int64             `json:"total"`
	Page    int               `json:"page"`
	PerPage int               `json:"per_page"`
	Results []TorrentMetadata `json:"results"`
}
