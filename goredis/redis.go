package goredis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/shopastro/logs"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type (
	GoRedisConfig struct {
		Type           string        `yaml:"type"`
		Host           string        `yaml:"host"`
		Port           int           `yaml:"port"`
		Db             int           `yaml:"db"`
		Username       string        `yaml:"username"`
		Password       string        `yaml:"password"`
		PoolSize       int           `yaml:"poolSize"`
		MaxConnAge     time.Duration `yaml:"maxConnAge"`
		IdleTimeout    time.Duration `yaml:"idleTimeout"`
		ConnectTimeout time.Duration `yaml:"connectTimeout"`
		ReadTimeout    time.Duration `yaml:"readTimeout"`
		WriteTimeout   time.Duration `yaml:"writeTimeout"`
		MasterName     string        `yaml:"masterName"`
		LogFile        string        `yaml:"logFile"`
	}
)

var (
	client        *redis.Client
	clusterClient *redis.ClusterClient
)

//func GetRedisClient(ctx context.Context) *redis.Client {
//	exp := &Exporter{}
//
//	return exp.WrapRedisClient(ctx, client)
//}

func GetClusterRedisClient() *redis.ClusterClient {
	return clusterClient
}

func NewClusterClient(cfg *GoRedisConfig) {
	clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MaxConnAge:   cfg.MaxConnAge * time.Millisecond,
		IdleTimeout:  cfg.IdleTimeout * time.Millisecond,
		DialTimeout:  cfg.ConnectTimeout * time.Millisecond,
		ReadTimeout:  cfg.ReadTimeout * time.Millisecond,
		WriteTimeout: cfg.ReadTimeout * time.Millisecond,
	})
}

func NewRedisClient(cfg *GoRedisConfig) *redis.Client {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	if cfg.Type == "sentinel" {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.MasterName,
			SentinelAddrs: []string{addr},
			Password:      cfg.Password,
			DB:            cfg.Db,
			PoolSize:      cfg.PoolSize,
			MaxConnAge:    cfg.MaxConnAge * time.Millisecond,
			IdleTimeout:   cfg.IdleTimeout * time.Millisecond,
			DialTimeout:   cfg.ConnectTimeout * time.Millisecond,
			ReadTimeout:   cfg.ReadTimeout * time.Millisecond,
			WriteTimeout:  cfg.ReadTimeout * time.Millisecond,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password:     cfg.Password,
			DB:           cfg.Db,
			PoolSize:     cfg.PoolSize,
			MaxConnAge:   cfg.MaxConnAge * time.Millisecond,
			IdleTimeout:  cfg.IdleTimeout * time.Millisecond,
			DialTimeout:  cfg.ConnectTimeout * time.Millisecond,
			ReadTimeout:  cfg.ReadTimeout * time.Millisecond,
			WriteTimeout: cfg.ReadTimeout * time.Millisecond,
		})
	}

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		logs.Logger.Fatal("[NewRedisClient]  error", zap.Error(err))
	}

	return client
}
