package config

import (
	"fmt"
	"os"
	"strings"

	platformconfig "github.com/amirhossein-shakeri/zhinux-platform/config"
)

const defaultServiceName = "zhinux-db-maintainer"

type Config struct {
	Base    platformconfig.Base
	Runtime platformconfig.Runtime
}

func Load() (Config, error) {
	if err := ensureServiceName(); err != nil {
		return Config{}, err
	}

	viperOptions := viperOptionsFromEnv()
	baseCfg, err := loadBase(viperOptions)
	if err != nil {
		return Config{}, err
	}

	runtimeCfg, err := loadRuntime(viperOptions)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Base:    baseCfg,
		Runtime: runtimeCfg,
	}
	if err := applyRedisOverrides(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, cfg.Runtime.Validate()
}

func ensureServiceName() error {
	if strings.TrimSpace(os.Getenv("SERVICE_NAME")) != "" {
		return nil
	}
	if err := os.Setenv("SERVICE_NAME", defaultServiceName); err != nil {
		return fmt.Errorf("set default service name: %w", err)
	}
	return nil
}

func loadBase(options platformconfig.ViperOptions) (platformconfig.Base, error) {
	cfg, err := platformconfig.LoadBaseWithViper(options)
	if err == nil {
		return cfg, nil
	}
	if strings.Contains(err.Error(), "viper support is disabled") {
		return platformconfig.LoadBaseFromEnv()
	}
	return platformconfig.Base{}, err
}

func loadRuntime(options platformconfig.ViperOptions) (platformconfig.Runtime, error) {
	cfg, err := platformconfig.LoadRuntimeWithViper(options)
	if err == nil {
		return cfg, nil
	}
	if strings.Contains(err.Error(), "viper support is disabled") {
		return platformconfig.LoadRuntimeFromEnv()
	}
	return platformconfig.Runtime{}, err
}

func applyRedisOverrides(cfg *Config) error {
	if cfg == nil {
		return nil
	}

	if overrideDB := strings.TrimSpace(os.Getenv("DB_MAINTAINER_REDIS_DATABASE")); overrideDB != "" {
		parsedDB, err := platformconfig.ParseOptionalInt(overrideDB)
		if err != nil {
			return fmt.Errorf("parse DB_MAINTAINER_REDIS_DATABASE: %w", err)
		}
		cfg.Runtime.Redis.Database = parsedDB
	}

	if overrideNamespace := strings.TrimSpace(os.Getenv("DB_MAINTAINER_REDIS_NAMESPACE")); overrideNamespace != "" {
		cfg.Runtime.Redis.Namespace = overrideNamespace
	}

	return nil
}

func viperOptionsFromEnv() platformconfig.ViperOptions {
	automaticEnv := true
	searchPaths := []string{"."}
	if rawPaths := strings.TrimSpace(os.Getenv("APP_CONFIG_PATHS")); rawPaths != "" {
		searchPaths = splitAndTrim(rawPaths)
	}

	return platformconfig.ViperOptions{
		ConfigFile:     strings.TrimSpace(os.Getenv("APP_CONFIG_FILE")),
		ConfigName:     withDefault(strings.TrimSpace(os.Getenv("APP_CONFIG_NAME")), "config"),
		ConfigType:     strings.TrimSpace(os.Getenv("APP_CONFIG_TYPE")),
		SearchPaths:    searchPaths,
		EnvPrefix:      strings.TrimSpace(os.Getenv("APP_CONFIG_ENV_PREFIX")),
		EnvKeyReplacer: strings.NewReplacer(".", "_"),
		AutomaticEnv:   &automaticEnv,
	}
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{"."}
	}
	return result
}

func withDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
