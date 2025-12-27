package notifier

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/alp/suumo-hunter-go/internal/models"
)

// mockHTTPClient is a mock implementation of HTTPClient for testing.
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
	}, nil
}

func TestCalculateScoreLabel(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  ScoreLabel
	}{
		{
			name:  "bargain",
			score: 15000,
			want:  ScoreLabelBargain,
		},
		{
			name:  "bargain threshold",
			score: 10000,
			want:  ScoreLabelBargain,
		},
		{
			name:  "standard positive",
			score: 5000,
			want:  ScoreLabelStandard,
		},
		{
			name:  "standard zero",
			score: 0,
			want:  ScoreLabelStandard,
		},
		{
			name:  "standard negative",
			score: -5000,
			want:  ScoreLabelStandard,
		},
		{
			name:  "expensive threshold",
			score: -10000,
			want:  ScoreLabelExpensive,
		},
		{
			name:  "expensive",
			score: -15000,
			want:  ScoreLabelExpensive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateScoreLabel(tt.score)
			if got != tt.want {
				t.Errorf("CalculateScoreLabel(%f) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestNotify(t *testing.T) {
	var capturedRequests []*http.Request
	var capturedBodies []string

	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			capturedRequests = append(capturedRequests, req)
			body, _ := io.ReadAll(req.Body)
			capturedBodies = append(capturedBodies, string(body))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		},
	}

	notifier := NewNotifier("test-token", WithHTTPClient(mock))
	ctx := context.Background()

	properties := []PropertyWithScore{
		{
			Property: models.Property{
				ID:            "jnc_001",
				Name:          "テストマンション",
				Address:       "東京都渋谷区",
				Rent:          79000,
				ManagementFee: 5000,
				URL:           "https://suumo.jp/chintai/jnc_001/",
			},
			Score: 12800,
			Label: ScoreLabelBargain,
		},
	}

	err := notifier.Notify(ctx, properties)
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}

	// Verify request was made
	if len(capturedRequests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(capturedRequests))
	}

	req := capturedRequests[0]

	// Verify authorization header
	auth := req.Header.Get("Authorization")
	if auth != "Bearer test-token" {
		t.Errorf("Authorization header = %q, want %q", auth, "Bearer test-token")
	}

	// Verify content type
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/x-www-form-urlencoded")
	}

	// Verify message content
	body := capturedBodies[0]
	if !strings.Contains(body, "message=") {
		t.Error("Request body should contain message parameter")
	}
	if !strings.Contains(body, "%E3%83%86%E3%82%B9%E3%83%88%E3%83%9E%E3%83%B3%E3%82%B7%E3%83%A7%E3%83%B3") {
		// URL encoded "テストマンション"
		t.Error("Request body should contain property name")
	}
}

func TestNotifyEmptyProperties(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			t.Error("Should not make request for empty properties")
			return nil, nil
		},
	}

	notifier := NewNotifier("test-token", WithHTTPClient(mock))
	ctx := context.Background()

	err := notifier.Notify(ctx, []PropertyWithScore{})
	if err != nil {
		t.Errorf("Notify() error = %v, expected nil", err)
	}
}

func TestNotifyWithManyProperties(t *testing.T) {
	requestCount := 0

	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			requestCount++
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		},
	}

	notifier := NewNotifier("test-token", WithHTTPClient(mock))
	ctx := context.Background()

	// Create 15 properties (more than MaxPropertiesPerNotification)
	properties := make([]PropertyWithScore, 15)
	for i := range 15 {
		properties[i] = PropertyWithScore{
			Property: models.Property{
				ID:            "jnc_" + string(rune('0'+i)),
				Name:          "マンション",
				Address:       "東京都",
				Rent:          80000,
				ManagementFee: 5000,
				URL:           "https://suumo.jp/",
			},
			Score: 0,
			Label: ScoreLabelStandard,
		}
	}

	err := notifier.Notify(ctx, properties)
	if err != nil {
		t.Fatalf("Notify() error = %v", err)
	}

	// Should have sent at least one request
	if requestCount == 0 {
		t.Error("Expected at least one request")
	}
}

func TestNotifyError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewReader([]byte("invalid token"))),
			}, nil
		},
	}

	notifier := NewNotifier("invalid-token", WithHTTPClient(mock))
	ctx := context.Background()

	properties := []PropertyWithScore{
		{
			Property: models.Property{
				Name: "Test",
				URL:  "https://example.com",
			},
			Label: ScoreLabelStandard,
		},
	}

	err := notifier.Notify(ctx, properties)
	if err == nil {
		t.Error("Expected error for unauthorized response")
	}
}

func TestFormatPropertyEntry(t *testing.T) {
	notifier := NewNotifier("test-token")

	tests := []struct {
		name     string
		prop     PropertyWithScore
		contains []string
	}{
		{
			name: "bargain property",
			prop: PropertyWithScore{
				Property: models.Property{
					Name:          "お得マンション",
					Address:       "東京都渋谷区",
					Rent:          79000,
					ManagementFee: 5000,
					URL:           "https://suumo.jp/test/",
				},
				Score: 12800,
				Label: ScoreLabelBargain,
			},
			contains: []string{"🏠", "お買い得", "お得マンション", "8.4万円", "12800円/月 お得"},
		},
		{
			name: "expensive property",
			prop: PropertyWithScore{
				Property: models.Property{
					Name:          "高いマンション",
					Address:       "東京都港区",
					Rent:          150000,
					ManagementFee: 10000,
					URL:           "https://suumo.jp/test/",
				},
				Score: -15000,
				Label: ScoreLabelExpensive,
			},
			contains: []string{"🏠", "割高", "高いマンション", "16.0万円", "15000円/月 高い"},
		},
		{
			name: "analyzing property",
			prop: PropertyWithScore{
				Property: models.Property{
					Name:          "分析中マンション",
					Address:       "東京都新宿区",
					Rent:          100000,
					ManagementFee: 5000,
					URL:           "https://suumo.jp/test/",
				},
				Score: 0,
				Label: ScoreLabelAnalyzing,
			},
			contains: []string{"🏠", "分析中"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := notifier.formatPropertyEntry(tt.prop)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("formatPropertyEntry() result should contain %q, got %q", expected, result)
				}
			}
		})
	}
}

func TestConvertToPropertyWithScore(t *testing.T) {
	properties := []models.Property{
		{ID: "jnc_001", Name: "Test 1"},
		{ID: "jnc_002", Name: "Test 2"},
	}

	result := ConvertToPropertyWithScore(properties)

	if len(result) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(result))
	}

	for i, p := range result {
		if p.Label != ScoreLabelAnalyzing {
			t.Errorf("Property[%d].Label = %v, want %v", i, p.Label, ScoreLabelAnalyzing)
		}
		if p.Score != 0 {
			t.Errorf("Property[%d].Score = %f, want 0", i, p.Score)
		}
	}
}

func TestGetLabelIcon(t *testing.T) {
	notifier := NewNotifier("test-token")

	// All labels should return the same unified icon
	labels := []ScoreLabel{
		ScoreLabelBargain,
		ScoreLabelExpensive,
		ScoreLabelAnalyzing,
		ScoreLabelStandard,
	}

	for _, label := range labels {
		t.Run(string(label), func(t *testing.T) {
			got := notifier.getLabelIcon(label)
			if got != "🏠" {
				t.Errorf("getLabelIcon(%v) = %q, want %q", label, got, "🏠")
			}
		})
	}
}

func TestMessageSplitting(t *testing.T) {
	notifier := NewNotifier("test-token")

	// Create properties with long names to force message splitting
	properties := make([]PropertyWithScore, 5)
	for i := range 5 {
		properties[i] = PropertyWithScore{
			Property: models.Property{
				Name:    strings.Repeat("あ", 100), // Long name
				Address: strings.Repeat("い", 50),
				Rent:    80000,
				URL:     "https://suumo.jp/" + strings.Repeat("x", 50),
			},
			Score: 5000,
			Label: ScoreLabelStandard,
		}
	}

	messages := notifier.formatMessages(properties)

	// Should have multiple messages due to length limit
	if len(messages) < 2 {
		t.Errorf("Expected multiple messages for long content, got %d", len(messages))
	}

	// Each message should be within limit
	for i, msg := range messages {
		if len(msg) > MaxMessageLength {
			t.Errorf("Message[%d] length %d exceeds limit %d", i, len(msg), MaxMessageLength)
		}
	}
}
