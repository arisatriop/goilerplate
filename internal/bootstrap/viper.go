package bootstrap

import (
	"fmt"
	"goilerplate/config"
	"os"
	"reflect"
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

	// Bind environment variables to struct fields
	// This is necessary because Unmarshal doesn't pick up environment variables
	// for keys that aren't already defined in the config file or as defaults.
	bindEnvs(v, config.Config{})

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct, %w", err))
	}

	return &cfg
}

// bindEnvs recursively binds environment variables for nested structs using reflection.
func bindEnvs(v *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := ifv.Type()
	for i := 0; i < ift.NumField(); i++ {
		field := ift.Field(i)
		tv := field.Tag.Get("mapstructure")
		if tv == "-" {
			continue
		}
		if tv == "" {
			tv = strings.ToLower(field.Name)
		}

		// Handle nested structs
		path := append(parts, tv)
		if field.Type.Kind() == reflect.Struct {
			bindEnvs(v, reflect.New(field.Type).Elem().Interface(), path...)
		} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			bindEnvs(v, reflect.New(field.Type.Elem()).Elem().Interface(), path...)
		} else if field.Type.Kind() == reflect.Map {
			// Scan environment variables for map keys
			envPrefix := strings.ToUpper(strings.Join(path, "_")) + "_"
			elemType := field.Type.Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}

			for _, env := range os.Environ() {
				if strings.HasPrefix(env, envPrefix) {
					pair := strings.SplitN(env, "=", 2)
					envVar := pair[0]

					if elemType.Kind() == reflect.Struct {
						// For map of structs, we need to find which field this env var matches
						// Example: SERVICE_GOPAY_NAME. envVar=SERVICE_GOPAY_NAME, envPrefix=SERVICE_
						// We check each field of Service struct.
						for j := 0; j < elemType.NumField(); j++ {
							f := elemType.Field(j)
							suffix := "_" + strings.ToUpper(f.Tag.Get("mapstructure"))
							if suffix == "_" {
								suffix = "_" + strings.ToUpper(f.Name)
							}

							if strings.HasSuffix(envVar, suffix) {
								mapKey := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(envVar, envPrefix), suffix))
								fieldKey := strings.ToLower(strings.TrimPrefix(suffix, "_"))
								viperKey := strings.Join(path, ".") + "." + mapKey + "." + fieldKey
								v.BindEnv(viperKey, envVar)
							}
						}
					} else {
						// Simple map (e.g. map[string]string)
						keyPart := strings.ToLower(strings.TrimPrefix(envVar, envPrefix))
						viperKey := strings.Join(path, ".") + "." + keyPart
						v.BindEnv(viperKey, envVar)
					}
				}
			}
		} else {
			key := strings.Join(path, ".")
			envVar := strings.ToUpper(strings.Join(path, "_"))
			v.BindEnv(key, envVar)
		}
	}
}
