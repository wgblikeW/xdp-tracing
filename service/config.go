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
	WriteTimeout    time.Duration `yaml:"writeTimeout"`
	ReadTimeout     time.Duration `yaml:"readTimeout"`
	DialTimeout     time.Duration `yaml:"dialTimeout"`
	MaxRetryBackoff time.Duration `yaml:"maxRetryBackoff"`
	MinRetryBackoff time.Duration `yaml:"minRetryBackoff"`
	MaxRetries      int           `yaml:"maxRetries"`
	Username        string        `yaml:"username"`
	Network         string        `yaml:"network"`
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	Db              int           `yaml:"db"`
}

func MakeNewRedisOptions() *redis.Options {
	redisConfig := extractRedisConfig()
	var redisOptions = &redis.Options{
		PoolFIFO:        redisConfig.PoolFIFO,
		WriteTimeout:    redisConfig.WriteTimeout,
		ReadTimeout:     redisConfig.ReadTimeout,
		DialTimeout:     redisConfig.DialTimeout,
		MaxRetries:      redisConfig.MaxRetries,
		MinRetryBackoff: redisConfig.MinRetryBackoff,
		Username:        redisConfig.Username,
		Network:         redisConfig.Network,
		Addr:            redisConfig.Network,
		Password:        redisConfig.Password,
		DB:              redisConfig.Db,
	}
	return redisOptions
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
