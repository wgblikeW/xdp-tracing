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
	Etcd         *EtcdConfig         `yaml:"etcd"`
	Grpc         *GrpcConfig         `yaml:"grpc"`
	Rest         *RestConfig         `yaml:"rest"`
}

var gConfig *Config

type stringList []string

type RestConfig struct {
	Addr       string `yaml:"addr"`
	Production bool   `yaml:"production"`
}

type GrpcConfig struct {
	Port  int    `yaml:"port"`
	MapID uint32 `yaml:"mapid"`
}

type EtcdConfig struct {
	EndPoints            stringList    `yaml:"endpoints"`
	Dialtimeout          time.Duration `yaml:"dial-timeout"`
	AutoSyncInterval     time.Duration `json:"auto-sync-interval"`
	DialKeepAliveTime    time.Duration `json:"dial-keep-alive-time"`
	DialKeepAliveTimeout time.Duration `json:"dial-keep-alive-timeout"`
	Username             string        `json:"username"`
	Password             string        `json:"password"`
	RejectOldCluster     bool          `json:"reject-old-cluster"`
	PermitWithoutStream  bool          `json:"permit-without-stream"`
}

type PacketFilterConfig struct {
	SrcIP   stringList `yaml:"srcip"`
	DstIP   stringList `yaml:"dstip"`
	SrcPort stringList `yaml:"srcport"`
	DstPort stringList `yaml:"dstport"`
}

// Part of the fields in redis.Options
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

// MakeNewRedisOptions read rules from given config file and parse it into
// Options field
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

// MakeNewRules read rules from given config file and parse it into
// Rules field
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
}

func extractgRPCConfig() *GrpcConfig {
	return gConfig.Grpc
}

func extractEtcdConfig() *EtcdConfig {
	return gConfig.Etcd
}

func extractRedisConfig() *RedisConfig {
	return gConfig.RedisDB
}

func extractPacketFilterConfig() *PacketFilterConfig {
	return gConfig.PacketFilter
}

func extractRestConfig() *RestConfig {
	return gConfig.Rest
}

func ExtractRestConfig() *RestConfig {
	return extractRestConfig()
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
