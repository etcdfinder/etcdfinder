package config

import (
	"strings"

	"github.com/etcdfinder/etcdfinder/internal/lib"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Log       LogConfig       `mapstructure:"log"`
	Etcd      EtcdConfig      `mapstructure:"etcd"`
	Datastore DatastoreConfig `mapstructure:"datastore"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type LogConfig struct {
	Level lib.LogLevel `mapstructure:"level"`
}

type DatastoreConfig struct {
	Type        string            `mapstructure:"type"`
	Meilisearch MeilisearchConfig `mapstructure:"meilisearch"`
}

type EtcdConfig struct {
	Endpoints             string `mapstructure:"endpoints"`
	RootPrefixEtcd        string `mapstructure:"root_etcd_prefix"`
	WatchEventChannelSize int    `mapstructure:"watch_event_channel_size"`
	PaginationLimit       int    `mapstructure:"pagination_limit"`
}

type MeilisearchConfig struct {
	Host             string `mapstructure:"host"`
	IndexName        string `mapstructure:"index_name"`
	MatchingStrategy string `mapstructure:"matching_strategy"`
}

func Load(configPath string) (*Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	// Environment variables will have highest priority
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
