package utils

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Minio struct {
		AccessKey string `envconfig:"MINIO_ACCESS_KEY"`
		SecretKey string `envconfig:"MINIO_SECRET_KEY"`
		Endpoint  string `envconfig:"MINIO_ENDPOINT"`
	}
	Web struct {
		ListenAddress string `envconfig:"WEB_LISTEN_ADDRESS"`
		Prefork       bool   `envconfig:"WEB_PREFORK"`
	}
	Server struct {
		Ip   string `envconfig:"SERVER_IP"`
		Port int    `envconfig:"SERVER_PORT"`
	}
	Gorse struct {
		Address string `envconfig:"GORSE_ADDRESS"`
		Key     string `envconfig:"GORSE_KEY"`
	}
	SnowflakeNode   int64
	AmqpUrl         string `envconfig:"AMQP_URL"`
	TranscodeNvidia bool   `envconfig:"TRANSCODE_NVIDIA"`
	PostgresUrl     string `envconfig:"POSTGRES_URL"`
	MediaAddress    string `envconfig:"MEDIA_ADDRESS"`
	CdnAddress      string `envconfig:"CDN_ADDRESS"`
}

func ParseConfig() (config Config) {
	envconfig.MustProcess("", &config)
	return
}
