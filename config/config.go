package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	Elasticsearch struct {
		Host  string `mapstructure:"host"`
		Index string `mapstructure:"index"`
	} `mapstructure:"elasticsearch"`
	Pagination struct {
		PageSize int `mapstructure:"page_size"`
	} `mapstructure:"pagination"`
}

var AppConfig Config

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("elasticsearch.host", "http://192.168.30.226:9200")
	viper.SetDefault("elasticsearch.index", "bittorrent_metadata")
	viper.SetDefault("pagination.page_size", 15)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("未找到配置文件，使用默认配置")
		} else {
			return err
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return err
	}

	return nil
}
