package bootstrap

import (
	"fmt"
	"goilerplate/config"

	"strings"

	"github.com/spf13/viper"
)

func Load() *config.Config {

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// SetEnvKeyReplacer allows mapping environment variables with underscores
	// to nested struct fields (e.g. SERVICE_0_NAME -> Service[0].Name)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// If the config file is not found, we don't panic.
		// This allows the app to run using ONLY environment variables (Pure Env).
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct, %w", err))
	}

	return &cfg
}
