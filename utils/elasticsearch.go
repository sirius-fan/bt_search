package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirius/bt_search/config"
	"github.com/sirius/bt_search/models"
)

type ESClient struct {
	client *elasticsearch.Client
}

var esClient *ESClient

// InitElasticsearch 初始化Elasticsearch客户端
func InitElasticsearch() error {
	cfg := elasticsearch.Config{
		Addresses: []string{config.AppConfig.Elasticsearch.Host},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}

	// 测试连接
	resp, err := client.Info()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("Elasticsearch连接错误: %s", resp.String())
	}

	esClient = &ESClient{client: client}
	log.Println("Elasticsearch连接成功")
	return nil
}

// SearchTorrents 从ES中搜索种子数据
func SearchTorrents(req models.SearchRequest) (*models.SearchResponse, error) {
	if esClient == nil {
		return nil, fmt.Errorf("Elasticsearch客户端未初始化")
	}

	pageSize := config.AppConfig.Pagination.PageSize
	from := (req.Page - 1) * pageSize

	// 构建搜索查询
	var query map[string]interface{}

	if req.Query == "" {
		// 如果没有查询关键词，则获取所有文档
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
			"from": from,
			"size": pageSize,
		}
	} else if strings.HasPrefix(req.Query, "info_hash:") {
		// 如果是info_hash精确查询
		infoHash := strings.TrimPrefix(req.Query, "info_hash:")
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]interface{}{
					"info_hash": infoHash,
				},
			},
			"from": from,
			"size": pageSize,
		}
	} else {
		// 如果有其他查询关键词，进行多字段搜索
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  req.Query,
					"fields": []string{"name^3", "files.path"}, // 名称权重更高
				},
			},
			"from": from,
			"size": pageSize,
		}
	}

	// 添加排序
	if req.SortBy != "" {
		var sortField string
		switch req.SortBy {
		case "date":
			sortField = "create_date"
		case "size":
			sortField = "total_size"
		case "files":
			sortField = "file_count"
		default:
			sortField = "_score" // 默认按相关性排序
		}

		sortOrder := "desc"
		if req.SortOrder == "asc" {
			sortOrder = "asc"
		}

		query["sort"] = []map[string]interface{}{
			{sortField: map[string]interface{}{"order": sortOrder}},
		}
	}

	// 执行搜索
	jsonQuery, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	res, err := esClient.client.Search(
		esClient.client.Search.WithContext(context.Background()),
		esClient.client.Search.WithIndex(config.AppConfig.Elasticsearch.Index),
		esClient.client.Search.WithBody(strings.NewReader(string(jsonQuery))),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("搜索错误: %s", res.String())
	}

	// 解析搜索结果
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 解析总数和结果列表
	total := int64(result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})

	// 构造响应
	response := &models.SearchResponse{
		Total:   total,
		Page:    req.Page,
		PerPage: pageSize,
		Results: make([]models.TorrentMetadata, 0, len(hits)),
	}

	// 遍历搜索结果并填充响应
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		sourceBytes, err := json.Marshal(source)
		if err != nil {
			log.Printf("解析结果错误: %v", err)
			continue
		}

		var torrent models.TorrentMetadata
		if err := json.Unmarshal(sourceBytes, &torrent); err != nil {
			log.Printf("解析种子元数据错误: %v", err)
			continue
		}

		response.Results = append(response.Results, torrent)
	}

	return response, nil
}
