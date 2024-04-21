package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/mqufflc/whodidthechores/internal/api"
	"github.com/mqufflc/whodidthechores/internal/repository"
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
}

func (c Config) Validate() error {
	if c.Port < 1024 || c.Port > 5000 {
		return errors.New("application port must be between 1024 and 5000")
	}
	if err := c.Database.Validate(); err != nil {
		return err
	}
	return nil
}

func main() {

	viperInstance := viper.New()
	viperInstance.SetConfigName("config")
	viperInstance.AddConfigPath("/etc/whodidthechores/")
	viperInstance.AddConfigPath(".")
	err := viperInstance.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No Config file found.")
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	viperInstance.SetEnvPrefix("WDTC")
	replacer := strings.NewReplacer(".", "_")
	viperInstance.SetEnvKeyReplacer(replacer)
	viperInstance.AutomaticEnv()

	viperInstance.SetDefault("port", 3000)
	viperInstance.SetDefault("database.username", "")
	viperInstance.SetDefault("database.password", "")
	viperInstance.SetDefault("database.hostname", "")
	viperInstance.SetDefault("database.database", "whodidthechores")
	viperInstance.SetDefault("database.port", 5432)
	viperInstance.SetDefault("database.sslMode", "disable")

	var config Config

	err = viperInstance.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("unable to decode config into struct, %w", err))
	}
	if err = config.Validate(); err != nil {
		panic(fmt.Errorf("configuration error: %w", err))
	}

	service, err := repository.NewService(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", config.Database.Username, config.Database.Password, config.Database.Hostname, config.Database.Port, config.Database.Database, config.Database.SslMode))
	if err != nil {
		log.Fatal(err)
	}
	err = service.Migrate("migrations")
	if err != nil {
		log.Fatal(err)
	}
	defaultHandler := api.New(service)

	http := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: defaultHandler,
	}

	fmt.Printf("Listening on :%d\n", config.Port)
	http.ListenAndServe()
}
