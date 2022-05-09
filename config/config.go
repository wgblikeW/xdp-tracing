package config

type Config interface {
	NewConfig() *Config
}
