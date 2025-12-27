// Package main is the entry point for the SUUMO Hunter Lambda function.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/alp/suumo-hunter/internal/analyzer"
	"github.com/alp/suumo-hunter/internal/config"
	"github.com/alp/suumo-hunter/internal/models"
	"github.com/alp/suumo-hunter/internal/notifier"
	"github.com/alp/suumo-hunter/internal/scraper"
	"github.com/alp/suumo-hunter/internal/storage"
)

func main() {
	lambda.Start(Handler)
}

// Handler is the Lambda function handler.
func Handler(ctx context.Context) error {
	log.Println("Starting SUUMO Hunter...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	log.Printf("Config loaded: bucket=%s, key=%s, maxPage=%d",
		cfg.BucketName, cfg.BucketKey, cfg.MaxPage)

	// Initialize AWS SDK
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)

	// Initialize components
	store := storage.NewStorage(s3Client, cfg.BucketName, cfg.BucketKey)
	scrp := scraper.NewScraper(cfg.SuumoSearchURL, scraper.WithMaxPages(cfg.MaxPage))
	notify := notifier.NewNotifier(cfg.DiscordWebhookURL)
	analyze := analyzer.NewAnalyzer()

	// Step 1: Download previous data from S3
	log.Println("Downloading previous data from S3...")
	previousProperties, err := store.Download(ctx)
	if err != nil {
		return fmt.Errorf("failed to download previous data: %w", err)
	}
	log.Printf("Previous properties: %d", len(previousProperties))

	// Step 2: Scrape SUUMO
	log.Printf("Scraping SUUMO (max %d pages)...", cfg.MaxPage)
	currentProperties, err := scrp.Scrape(ctx)
	if err != nil {
		return fmt.Errorf("failed to scrape SUUMO: %w", err)
	}
	log.Printf("Current properties: %d", len(currentProperties))

	// Step 3: Find new properties
	newProperties := models.FindNewProperties(currentProperties, previousProperties)
	log.Printf("New properties: %d", len(newProperties))

	// Step 4: Merge and save to S3
	mergedProperties := models.MergeProperties(currentProperties, previousProperties)
	log.Printf("Uploading merged data (%d properties) to S3...", len(mergedProperties))
	if err := store.Upload(ctx, mergedProperties); err != nil {
		return fmt.Errorf("failed to upload data: %w", err)
	}

	// Step 5: Analyze and notify if there are new properties
	if len(newProperties) > 0 {
		log.Println("Running regression analysis...")
		// Use merged data for regression, but only analyze new properties
		scoredProperties := analyze.AnalyzeNewProperties(mergedProperties, newProperties)

		log.Println("Sending Discord notification...")
		if err := notify.Notify(ctx, scoredProperties); err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}
		log.Printf("Notified %d new properties", len(newProperties))
	} else {
		log.Println("No new properties found, skipping notification")
	}

	log.Println("SUUMO Hunter completed successfully!")
	return nil
}
