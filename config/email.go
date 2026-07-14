package config

import (
	"github.com/spf13/viper"
)

type EmailConfig struct {
	SenderEmail				string				`mapstructure:"BREVO_SENDER_EMAIL"`
	SenderName				string				`mapstructure:"BREVO_SENDER_NAME"`
	APIKey						string				`mapstructure:"BREVO_API_KEY"`
}

func NewEmailConfig() (*EmailConfig, error) {
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.AutomaticEnv()

	var config EmailConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}