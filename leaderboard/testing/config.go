package testing

import (
	"context"
	"strings"

	"github.com/spf13/viper"
)

//Creates an empty context (shortcut for context.Background())
func NewEmptyCtx() context.Context {
	return context.Background()
}

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
