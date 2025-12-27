// Package config provides configuration management for the SUUMO Hunter.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	// BucketName is the S3 bucket name for storing property data.
	BucketName string `env:"BUCKET_NAME,required"`

	// BucketKey is the S3 object key for the CSV file.
	BucketKey string `env:"BUCKET_KEY" envDefault:"properties.csv"`

	// MaxPage is the maximum number of SUUMO pages to scrape.
	MaxPage int `env:"MAX_PAGE" envDefault:"30"`

	// SuumoSearchURL is the SUUMO search result URL to scrape.
	SuumoSearchURL string `env:"SUUMO_SEARCH_URL,required"`

	// DiscordWebhookURL is the Discord Webhook URL for notifications.
	DiscordWebhookURL string `env:"DISCORD_WEBHOOK_URL,required"`
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}
