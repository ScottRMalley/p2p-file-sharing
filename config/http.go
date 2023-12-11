package config

import "github.com/kelseyhightower/envconfig"

type HttpEnv struct {
	Port  int  `split_words:"true" required:"true" default:"8080"`
	Debug bool `split_words:"true" required:"true" default:"false"`
}

func ParseHttpEnv(prefix string) HttpEnv {
	var httpConfig HttpEnv
	if err := envconfig.Process(prefix, &httpConfig); err != nil {
		panic(err)
	}
	return httpConfig
}
