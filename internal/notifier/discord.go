// Package notifier provides Discord Webhook integration for sending property alerts.
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alp/suumo-hunter-go/internal/models"
)

const (
	// MaxMessageLength is the maximum length of a single Discord message.
	MaxMessageLength = 2000

	// MaxPropertiesPerNotification is the maximum number of properties to include in one notification.
	MaxPropertiesPerNotification = 10

	// BargainThreshold is the threshold (in yen) for considering a property a bargain.
	BargainThreshold = 10000

	// ExpensiveThreshold is the threshold (in yen) for considering a property expensive.
	ExpensiveThreshold = -10000
)

// ScoreLabel represents the bargain level of a property.
type ScoreLabel string

const (
	ScoreLabelBargain   ScoreLabel = "お買い得"
	ScoreLabelStandard  ScoreLabel = "標準"
	ScoreLabelExpensive ScoreLabel = "割高"
	ScoreLabelAnalyzing ScoreLabel = "分析中"
)

// PropertyWithScore represents a property with its bargain score.
type PropertyWithScore struct {
	Property models.Property
	Score    float64    // Bargain score in yen (positive = cheaper than expected)
	Label    ScoreLabel // Score label
}

// CalculateScoreLabel determines the score label based on the score value.
func CalculateScoreLabel(score float64) ScoreLabel {
	if score >= BargainThreshold {
		return ScoreLabelBargain
	}
	if score <= ExpensiveThreshold {
		return ScoreLabelExpensive
	}
	return ScoreLabelStandard
}

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// discordPayload is the JSON payload for Discord Webhook.
type discordPayload struct {
	Content string `json:"content"`
}

// Notifier sends notifications via Discord Webhook.
type Notifier struct {
	webhookURL string
	client     HTTPClient
}

// Option is a function that configures a Notifier.
type Option func(*Notifier)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c HTTPClient) Option {
	return func(n *Notifier) {
		n.client = c
	}
}

// NewNotifier creates a new Notifier with the given Discord webhook URL.
func NewNotifier(webhookURL string, opts ...Option) *Notifier {
	n := &Notifier{
		webhookURL: webhookURL,
		client:     http.DefaultClient,
	}

	for _, opt := range opts {
		opt(n)
	}

	return n
}

// Notify sends a notification for new properties.
// If there are more than MaxPropertiesPerNotification properties,
// only the first MaxPropertiesPerNotification are shown with a summary.
func (n *Notifier) Notify(ctx context.Context, properties []PropertyWithScore) error {
	if len(properties) == 0 {
		return nil
	}

	messages := n.formatMessages(properties)

	for _, msg := range messages {
		if err := n.send(ctx, msg); err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}
	}

	return nil
}

// formatMessages creates notification messages from properties.
// Messages are split if they exceed MaxMessageLength.
func (n *Notifier) formatMessages(properties []PropertyWithScore) []string {
	var messages []string
	var currentMsg strings.Builder

	// Header
	currentMsg.WriteString("🏠 **新着物件のお知らせ**\n")

	// Limit to MaxPropertiesPerNotification
	displayProps := properties
	remaining := 0
	if len(properties) > MaxPropertiesPerNotification {
		displayProps = properties[:MaxPropertiesPerNotification]
		remaining = len(properties) - MaxPropertiesPerNotification
	}

	for _, prop := range displayProps {
		entry := n.formatPropertyEntry(prop)

		// Check if adding this entry would exceed the limit
		if currentMsg.Len()+len(entry) > MaxMessageLength {
			// Save current message and start a new one
			messages = append(messages, currentMsg.String())
			currentMsg.Reset()
			currentMsg.WriteString("🏠 **新着物件のお知らせ（続き）**\n")
		}

		currentMsg.WriteString(entry)
	}

	// Add remaining count if any
	if remaining > 0 {
		summary := fmt.Sprintf("\n📋 他%d件の新着あり\n", remaining)
		if currentMsg.Len()+len(summary) > MaxMessageLength {
			messages = append(messages, currentMsg.String())
			currentMsg.Reset()
		}
		currentMsg.WriteString(summary)
	}

	// Add final message
	if currentMsg.Len() > 0 {
		messages = append(messages, currentMsg.String())
	}

	return messages
}

// formatPropertyEntry formats a single property for notification.
func (n *Notifier) formatPropertyEntry(prop PropertyWithScore) string {
	var sb strings.Builder

	// Property name
	sb.WriteString(fmt.Sprintf("\n**■ %s**\n", prop.Property.Name))

	// Address
	sb.WriteString(fmt.Sprintf("📍 %s\n", prop.Property.Address))

	// Total rent in 万円
	totalRent := prop.Property.TotalRentMan()
	sb.WriteString(fmt.Sprintf("💰 %.1f万円（管理費込）\n", totalRent))

	// Score
	if prop.Label != ScoreLabelAnalyzing {
		if prop.Score >= 0 {
			sb.WriteString(fmt.Sprintf("💴 相場より %.0f円/月 お得\n", prop.Score))
		} else {
			sb.WriteString(fmt.Sprintf("💴 相場より %.0f円/月 高い\n", -prop.Score))
		}
	}

	// URL
	sb.WriteString(fmt.Sprintf("🔗 %s\n", prop.Property.URL))

	return sb.String()
}

// send sends a message to Discord Webhook.
func (n *Notifier) send(ctx context.Context, message string) error {
	payload := discordPayload{Content: message}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Discord returns 204 No Content on success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Discord Webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ConvertToPropertyWithScore converts properties to PropertyWithScore with analyzing label.
// Use this when regression analysis is not available.
func ConvertToPropertyWithScore(properties []models.Property) []PropertyWithScore {
	result := make([]PropertyWithScore, len(properties))
	for i, p := range properties {
		result[i] = PropertyWithScore{
			Property: p,
			Score:    0,
			Label:    ScoreLabelAnalyzing,
		}
	}
	return result
}
