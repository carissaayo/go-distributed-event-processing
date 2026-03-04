package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig
	MongoDB    MongoDBConfig
	Redis      RedisConfig
	Processing ProcessingConfig
	Logging    LoggingConfig
}

type ServerConfig struct {
	Port         int
	Env          string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type MongoDBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

type RedisConfig struct {
	URL string
}

type ProcessingConfig struct {
	WorkerCount   int
	BufferSize    int
	BatchSize     int
	FlushInterval time.Duration
}

type LoggingConfig struct {
	Level  string
	Format string
}