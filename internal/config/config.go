package config

import (
	"github.com/spf13/viper"
)

// Config stores configuration data
type Config struct {
	DBHost        string `mapstructure:"DB_HOST"`
	DBPort        string `mapstructure:"DB_PORT"`
	DBUsername    string `mapstructure:"DB_USERNAME"`
	DBPassword    string `mapstructure:"DB_PASSWORD"`
	DBName        string `mapstructure:"DB_NAME"`
	RabbitMQHost  string `mapstructure:"RABBITMQ_HOST"`
	RabbitMQVHost string `mapstructure:"RABBITMQ_VHOST"`
	RabbitMQPort  string `mapstructure:"RABBITMQ_PORT"`
	RabbitMQUser  string `mapstructure:"RABBITMQ_USERNAME"`
	RabbitMQPass  string `mapstructure:"RABBITMQ_PASSWORD"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
