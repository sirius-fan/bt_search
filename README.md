# 磁力链接搜索网站

这是一个基于Go语言和Elasticsearch的磁力链接搜索系统，提供了简洁的界面和高效的搜索功能。

## 功能特点

- 首页展示最新的15个磁力链接
- 支持关键字搜索磁力链接
- 支持分页显示搜索结果
- 支持按不同维度排序：
  - 上传时间（最新/最早）
  - 文件大小（从大到小/从小到大）
  - 文件数量（从多到少/从少到多）
- 详情页展示种子的详细信息，包括文件列表
- 支持复制和使用磁力链接

## 技术栈

- 后端：Go + Gin框架
- 数据存储：Elasticsearch
- 前端：HTML + CSS + JavaScript + Bootstrap 5

## 系统要求

- Go 1.18+
- Elasticsearch 8.x
- Docker（可选，用于容器化部署）

## 安装部署

### 本地开发环境

1. 克隆项目到本地

```bash
git clone https://github.com/yourusername/bt_search.git
cd bt_search
```

2. 安装依赖

```bash
go mod tidy
```

3. 启动Elasticsearch（确保已安装或使用Docker）

```bash
docker run -d --name elasticsearch -p 9200:9200 -p 9300:9300 \
  -e "discovery.type=single-node" -e "xpack.security.enabled=false" \
  docker.elastic.co/elasticsearch/elasticsearch:8.6.0
```

4. 创建Elasticsearch索引

```bash
curl -X PUT "http://localhost:9200/bittorrent_metadata" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1
  },
  "mappings": {
    "properties": {
      "info_hash": {
        "type": "keyword"
      },
      "name": {
        "type": "text",
        "analyzer": "standard"
      },
      "files": {
        "type": "nested",
        "properties": {
          "path": {
            "type": "text",
            "analyzer": "standard"
          },
          "length": {
            "type": "long"
          }
        }
      },
      "total_size": {
        "type": "long"
      },
      "file_count": {
        "type": "integer"
      },
      "create_date": {
        "type": "date"
      }
    }
  }
}'
```

5. 生成测试数据（可选）

```bash
cd tools
go run gen_sample_data.go
cd ..
```

6. 启动应用

```bash
go run main.go
```

7. 访问应用

打开浏览器，访问 http://localhost:8080

### 使用Docker Compose部署

1. 使用Docker Compose启动整个应用

```bash
docker-compose up -d
```

2. 访问应用

打开浏览器，访问 http://localhost:8080

## 配置说明

配置文件位于 `config/config.yaml`，可以修改以下配置项：

```yaml
server:
  port: "8080"
elasticsearch:
  host: "http://localhost:9200"
  index: "bittorrent_metadata"
pagination:
  page_size: 15
```

## 项目结构

```
bt_search/
├── api/               # API处理函数
├── config/            # 配置文件和配置加载
├── models/            # 数据模型
├── static/            # 静态资源（CSS、JS等）
├── templates/         # HTML模板
├── tools/             # 辅助工具
├── utils/             # 工具函数和服务
├── main.go            # 应用入口
├── go.mod             # Go模块文件
├── go.sum             # Go依赖校验文件
├── Dockerfile         # Docker构建文件
└── docker-compose.yml # Docker Compose配置文件
```

## 许可证

MIT License
