package service

import (
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

const (
	REDISFIELD = "redisdb"
)

type Config struct {
	RedisDB *RedisConfig `yaml:"redisdb"`
}

var gConfig *Config

func init() {
	gConfig = readAndParseConfig(os.Args[1])
}

// Some fields Mapping to redis.Options
type RedisConfig struct {
	PoolFIFO        bool          `yaml:"poolFIFO"`
	PoolSize        int           `yaml:"poolsize"`
	MaxRetries      int           `yaml:"maxRetries"`
	Db              int           `yaml:"db"`
	Username        string        `yaml:"username"`
	Network         string        `yaml:"network"`
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	WriteTimeout    time.Duration `yaml:"writeTimeout"`
	ReadTimeout     time.Duration `yaml:"readTimeout"`
	DialTimeout     time.Duration `yaml:"dialTimeout"`
	MaxRetryBackoff time.Duration `yaml:"maxRetryBackoff"`
	MinRetryBackoff time.Duration `yaml:"minRetryBackoff"`
}

func (redisService *RedisService) MakeNewRedisOptions() {
	redisConfig := extractRedisConfig()

	redisService.Options = &redis.Options{
		PoolSize:        redisConfig.PoolSize,
		PoolFIFO:        redisConfig.PoolFIFO,
		WriteTimeout:    redisConfig.WriteTimeout,
		ReadTimeout:     redisConfig.ReadTimeout,
		DialTimeout:     redisConfig.DialTimeout,
		MaxRetries:      redisConfig.MaxRetries,
		MinRetryBackoff: redisConfig.MinRetryBackoff,
		Username:        redisConfig.Username,
		Network:         redisConfig.Network,
		Addr:            redisConfig.Addr,
		Password:        redisConfig.Password,
		DB:              redisConfig.Db,
	}
}

func extractRedisConfig() *RedisConfig {
	return gConfig.RedisDB
}

func readAndParseConfig(filePath string) *Config {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(filePath)

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
	var _config *Config
	err = viper.Unmarshal(&_config)
	if err != nil {
		fmt.Println(err.Error())
	}
	return _config
}
