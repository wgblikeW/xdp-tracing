package service

import (
	"reflect"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	REDISFIELD = "redisdb"
)

type Config struct {
	RedisDB      *RedisConfig        `yaml:"redisdb"`
	PacketFilter *PacketFilterConfig `yaml:"packetfilter"`
}

var gConfig *Config

type stringList []string

type PacketFilterConfig struct {
	SrcIP   stringList `yaml:"srcip"`
	DstIP   stringList `yaml:"dstip"`
	SrcPort stringList `yaml:"srcport"`
	DstPort stringList `yaml:"dstport"`
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
	logrus.Debug("In MakeNewRedisOptions:66 RedisConfig:%v", redisConfig)
}

func (capturer *TCP_IPCapturer) MakeNewRules() {
	filterRules := extractPacketFilterConfig()
	logrus.Debug("In MakeNewRules:77 FilterRules:%v", filterRules)
	rules := make(map[string][]string)
	v := reflect.ValueOf(filterRules).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Len() == 0 {
			continue
		}
		rules[v.Type().Field(i).Name] = v.Field(i).Interface().(stringList)
	}
	capturer.Rules = rules
	logrus.Debug("In MakeNewRules:77 Rules:%v", rules)
}

func extractRedisConfig() *RedisConfig {
	return gConfig.RedisDB
}

func extractPacketFilterConfig() *PacketFilterConfig {
	return gConfig.PacketFilter
}

func ReadAndParseConfig(filePath string) error {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(filePath)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&gConfig)
	if err != nil {
		return err
	}
	return nil
}
