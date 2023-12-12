package config

import "github.com/kelseyhightower/envconfig"

type ClientEnv struct {
	N     int      `split_words:"true" required:"true" default:"1000"`
	Hosts []string `split_words:"true" required:"true" default:"http://localhost:8080,http://localhost:8081,http://localhost:8082"`
}

func ParseClientEnv(prefix string) ClientEnv {
	var clientConfig ClientEnv
	if err := envconfig.Process(prefix, &clientConfig); err != nil {
		panic(err)
	}
	return clientConfig
}
