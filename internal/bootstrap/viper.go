package bootstrap

import (
	"fmt"
	"goilerplate/config"

	"github.com/spf13/viper"
)

func Load() *config.Config {

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct, %w", err))
	}

	return &cfg
}
