package config

import (
	"time"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Logger LoggerConfig `yaml:"logger"`
	SQLite SQLiteConfig `yaml:"sqlite"`
	Redis  RedisConfig  `yaml:"redis"`
}

type ServerConfig struct {
	Host         string        `yaml:"host" env-default:"localhost"`
	Port         string        `yaml:"port" env-default:"5000"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"5s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"10s"`
}

type LoggerConfig struct {
	Mode string `yaml:"mode" env-default:"local"` // local | dev | prod
	DisableSrc bool `yaml:"disable_src" env-default:"false"`
	FilePath string `yaml:"file_path" env-default:""` // если пусто - вывод в stdout
}

type SQLiteConfig struct {
	Path string `yaml:"path" env-requred:"true"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr" env-default:"localhost:6379"`
	Password string `yaml:"password" env-required:"true"`
	Username string `yaml:"username" env-required:"true"`
}
