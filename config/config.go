package config

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/topfreegames/podium/leaderboard/v2/enriching"
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

type (
	PodiumConfig struct {
		Enrichment enriching.EnrichmentConfig
	}
)

func DecodeHook() viper.DecoderConfigOption {
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		StringToMapStringHookFunc(),
		StringToMapBoolHookFunc(),
	)

	return viper.DecodeHook(decodeHook)

}

func StringToMapStringHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Map {
			return data, nil
		}

		if t.Key().Kind() != reflect.String || t.Elem().Kind() != reflect.String {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return map[string]string{}, nil
		}

		m := map[string]string{}
		err := json.Unmarshal([]byte(raw), &m)
		return m, err
	}
}

func StringToMapBoolHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Map {
			return data, nil
		}

		if t.Key().Kind() != reflect.String || t.Elem().Kind() != reflect.Bool {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return map[string]bool{}, nil
		}

		unmarshalled := map[string]string{}
		err := json.Unmarshal([]byte(raw), &unmarshalled)
		if err != nil {
			return map[string]bool{}, err
		}

		result := map[string]bool{}
		for k, v := range unmarshalled {
			boolValue, err := strconv.ParseBool(v)
			if err != nil {
				continue
			}
			result[k] = boolValue
		}

		return result, err
	}
}
