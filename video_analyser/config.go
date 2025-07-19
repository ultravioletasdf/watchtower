package main

import "github.com/kelseyhightower/envconfig"

type Config struct {
	AmqpUrl string `envconfig:"AMQP_URL"`

	Minio struct {
		AccessKey string `envconfig:"MINIO_ACCESS_KEY"`
		SecretKey string `envconfig:"MINIO_SECRET_KEY"`
		Endpoint  string `envconfig:"MINIO_ENDPOINT"`
	}
}

var config Config

func parseConfig() {
	envconfig.MustProcess("", &config)
}
