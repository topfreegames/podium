package config

import (
	"strings"

	"github.com/spf13/viper"
)

// GetDefaultConfig configure viper to use the config file
func GetDefaultConfig(configFile string) (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigFile(configFile)
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.SetEnvPrefix("podium")
	config.AddConfigPath("$HOME")
	config.AutomaticEnv()

	err := config.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}
