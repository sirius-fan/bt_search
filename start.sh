#!/bin/bash

# 检查配置文件是否存在
if [ ! -f ./config/config.yaml ]; then
    echo "配置文件不存在，创建默认配置..."
    mkdir -p ./config
    cat > ./config/config.yaml << EOF
server:
  port: "8080"
elasticsearch:
  host: "http://localhost:9200"
  index: "bittorrent_metadata"
pagination:
  page_size: 15
EOF
fi

# 检查Elasticsearch是否可访问
echo "正在检查Elasticsearch连接..."
curl -s --connect-timeout 5 http://localhost:9200 > /dev/null
if [ $? -ne 0 ]; then
    echo "无法连接到Elasticsearch，请确保Elasticsearch已启动"
    echo "可以使用以下命令启动Elasticsearch容器："
    echo "docker run -d --name elasticsearch -p 9200:9200 -p 9300:9300 -e \"discovery.type=single-node\" -e \"xpack.security.enabled=false\" docker.elastic.co/elasticsearch/elasticsearch:8.6.0"
    exit 1
fi

# 检查索引是否存在
INDEX_NAME=$(grep "index:" ./config/config.yaml | awk '{print $2}' | tr -d '"')
CHECK_INDEX=$(curl -s http://localhost:9200/${INDEX_NAME})
if echo $CHECK_INDEX | grep -q "index_not_found_exception"; then
    echo "未找到索引 ${INDEX_NAME}，正在创建..."
    curl -X PUT "http://localhost:9200/${INDEX_NAME}" -H 'Content-Type: application/json' -d'
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
    
    echo "索引已创建，要生成测试数据吗？(y/n)"
    read -r generate_data
    if [[ $generate_data == "y" || $generate_data == "Y" ]]; then
        echo "正在生成测试数据..."
        if [ -f ./tools/gen_sample_data ]; then
            ./tools/gen_sample_data
        else
            go run ./tools/gen_sample_data.go
        fi
    fi
fi

# 启动应用
echo "正在启动应用..."
go run main.go
