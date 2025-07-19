package main

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Minio struct {
		AccessKey string `envconfig:"MINIO_ACCESS_KEY"`
		SecretKey string `envconfig:"MINIO_SECRET_KEY"`
		Endpoint  string `envconfig:"MINIO_ENDPOINT"`
	}
	Port          int
	SnowflakeNode int64
	AmqpUrl       string `envconfig:"AMQP_URL"`
}

var config Config

func parseConfig() {
	envconfig.MustProcess("", &config)
}
