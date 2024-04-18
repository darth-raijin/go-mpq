package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config stores configuration data
type Config struct {
	DBHost       string `mapstructure:"DB_HOST"`
	DBPort       string `mapstructure:"DB_PORT"`
	DBUser       string `mapstructure:"DB_USER"`
	DBPass       string `mapstructure:"DB_PASS"`
	RabbitMQHost string `mapstructure:"RABBITMQ_HOST"`
	RabbitMQPort string `mapstructure:"RABBITMQ_PORT"`
	RabbitMQUser string `mapstructure:"RABBITMQ_USER"`
	RabbitMQPass string `mapstructure:"RABBITMQ_PASS"`
}

func LoadConfig() (Config, error) {
	viper.AutomaticEnv()

	viper.SetEnvPrefix("GO_MQP")
	err := viper.BindEnv("DB_HOST")
	if err != nil {
		panic(fmt.Sprintf("Error binding DB_HOST: %s", err))
	}

	err = viper.BindEnv("DB_PORT")
	if err != nil {
		panic(fmt.Sprintf("Error binding DB_PORT: %s", err))
	}

	err = viper.BindEnv("DB_USER")
	if err != nil {
		panic(fmt.Sprintf("Error binding DB_USER: %s", err))
	}

	err = viper.BindEnv("DB_PASS")
	if err != nil {
		panic(fmt.Sprintf("Error binding DB_PASS: %s", err))
	}

	err = viper.BindEnv("RABBITMQ_HOST")
	if err != nil {
		panic(fmt.Sprintf("Error binding RABBITMQ_HOST: %s", err))
	}

	err = viper.BindEnv("RABBITMQ_PORT")
	if err != nil {
		panic(fmt.Sprintf("Error binding RABBITMQ_PORT: %s", err))
	}
	err = viper.BindEnv("RABBITMQ_USER")
	if err != nil {
		panic(fmt.Sprintf("Error binding RABBITMQ_USER: %s", err))
	}
	err = viper.BindEnv("RABBITMQ_PASS")
	if err != nil {
		panic(fmt.Sprintf("Error binding RABBITMQ_PASS: %s", err))
	}

	// Unmarshal the environment variables into the Config struct
	err = viper.Unmarshal(&config)
	return nil
}
