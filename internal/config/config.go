package config

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type DbConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Hostname string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	SslMode  string `mapstructure:"sslmode"`
}

func (c DbConfig) Validate() error {
	if c.Username == "" {
		return errors.New("database username is required")
	}
	if c.Password == "" {
		return errors.New("database password is required")
	}
	if c.Hostname == "" {
		return errors.New("database hostname is required")
	}
	validDbSslMode := []string{"disable", "allow", "prefer"}
	if !slices.Contains(validDbSslMode, c.SslMode) {
		return errors.New("only 'disable', 'allow' or 'prefer' are supported for postgres ssl mode")
	}
	return nil
}

type Config struct {
	Port     int      `mapstructure:"port"`
	Database DbConfig `mapstructure:"database"`
	TimeZone string   `mapstructure:"timezone"`
}

func (c *Config) Validate() error {
	if c.Port < 1024 || c.Port > 5000 {
		return errors.New("application port must be between 1024 and 5000")
	}
	if err := c.Database.Validate(); err != nil {
		return err
	}
	_, err := time.LoadLocation(c.TimeZone)
	if err != nil {
		slog.Error(fmt.Sprintf("Unrecognized time zone: %v, UTC will be used instead", c.TimeZone))
		c.TimeZone = "UTC"
	}
	return nil
}

func New() (Config, error) {
	config := Config{}
	viperInstance := viper.New()
	viperInstance.SetConfigName("config")
	viperInstance.AddConfigPath("/etc/whodidthechores/")
	viperInstance.AddConfigPath(".")
	err := viperInstance.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No Config file found.")
		} else {
			return config, fmt.Errorf("config file error: %w", err)
		}
	}

	viperInstance.SetEnvPrefix("WDTC")
	replacer := strings.NewReplacer(".", "_")
	viperInstance.SetEnvKeyReplacer(replacer)
	viperInstance.AutomaticEnv()

	viperInstance.SetDefault("port", 8080)
	viperInstance.SetDefault("timezone", "UTC")
	viperInstance.SetDefault("database.username", "")
	viperInstance.SetDefault("database.password", "")
	viperInstance.SetDefault("database.hostname", "")
	viperInstance.SetDefault("database.database", "whodidthechores")
	viperInstance.SetDefault("database.port", 5432)
	viperInstance.SetDefault("database.sslMode", "disable")

	err = viperInstance.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("unable to decode config into struct, %w", err)
	}
	if err = config.Validate(); err != nil {
		return config, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}
