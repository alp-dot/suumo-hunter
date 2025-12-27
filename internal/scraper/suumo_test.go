package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// sampleHTML is a simplified version of SUUMO's property listing HTML
const sampleHTML = `
<!DOCTYPE html>
<html>
<head><title>SUUMO</title></head>
<body>
<div class="cassetteitem">
	<div class="cassetteitem_content-title">テストマンション</div>
	<ul class="cassetteitem_detail">
		<li class="cassetteitem_detail-col1">東京都渋谷区渋谷1-1-1</li>
		<li class="cassetteitem_detail-col2">
			<div class="cassetteitem_detail-text">西武新宿線/新井薬師前駅 歩8分</div>
			<div class="cassetteitem_detail-text">西武新宿線/沼袋駅 歩10分</div>
		</li>
		<li class="cassetteitem_detail-col3">
			<div>築5年</div>
			<div>3階建</div>
		</li>
	</ul>
	<table class="cassetteitem_other">
		<tbody>
			<tr>
				<td>1</td>
				<td>-</td>
				<td>3階</td>
				<td><span class="cassetteitem_price--rent">7.9万円</span></td>
				<td><span class="cassetteitem_price--administration">5000円</span></td>
				<td><span class="cassetteitem_price--deposit">1ヶ月</span></td>
				<td><span class="cassetteitem_price--gratuity">1ヶ月</span></td>
				<td><span class="cassetteitem_madori">1K</span></td>
				<td><span class="cassetteitem_menseki">25.5m²</span></td>
				<td><a href="/chintai/jnc_000102396492/">詳細を見る</a></td>
			</tr>
			<tr>
				<td>2</td>
				<td>-</td>
				<td>2階</td>
				<td><span class="cassetteitem_price--rent">7.5万円</span></td>
				<td><span class="cassetteitem_price--administration">5000円</span></td>
				<td><span class="cassetteitem_price--deposit">-</span></td>
				<td><span class="cassetteitem_price--gratuity">1ヶ月</span></td>
				<td><span class="cassetteitem_madori">1R</span></td>
				<td><span class="cassetteitem_menseki">20.0m²</span></td>
				<td><a href="/chintai/jnc_000102396493/">詳細を見る</a></td>
			</tr>
		</tbody>
	</table>
</div>
<div class="cassetteitem">
	<div class="cassetteitem_content-title">テストアパート</div>
	<ul class="cassetteitem_detail">
		<li class="cassetteitem_detail-col1">東京都新宿区新宿2-2-2</li>
		<li class="cassetteitem_detail-col2">
			<div class="cassetteitem_detail-text">JR山手線/新宿駅 歩5分</div>
		</li>
		<li class="cassetteitem_detail-col3">
			<div>新築</div>
			<div>2階建</div>
		</li>
	</ul>
	<table class="cassetteitem_other">
		<tbody>
			<tr>
				<td>1</td>
				<td>-</td>
				<td>1階</td>
				<td><span class="cassetteitem_price--rent">10万円</span></td>
				<td><span class="cassetteitem_price--administration">-</span></td>
				<td><span class="cassetteitem_price--deposit">-</span></td>
				<td><span class="cassetteitem_price--gratuity">-</span></td>
				<td><span class="cassetteitem_madori">1LDK</span></td>
				<td><span class="cassetteitem_menseki">35.0m²</span></td>
				<td><a href="/chintai/jnc_000102396494/">詳細を見る</a></td>
			</tr>
		</tbody>
	</table>
</div>
<div class="pagination">
	<a href="?page=1">1</a>
	<a href="?page=2">2</a>
	<a href="?page=2">次へ</a>
</div>
</body>
</html>
`

const sampleHTMLNoNext = `
<!DOCTYPE html>
<html>
<head><title>SUUMO</title></head>
<body>
<div class="cassetteitem">
	<div class="cassetteitem_content-title">最後のマンション</div>
	<ul class="cassetteitem_detail">
		<li class="cassetteitem_detail-col1">東京都目黒区</li>
		<li class="cassetteitem_detail-col2">
			<div class="cassetteitem_detail-text">東急東横線/中目黒駅 歩3分</div>
		</li>
		<li class="cassetteitem_detail-col3">
			<div>築10年</div>
			<div>5階建</div>
		</li>
	</ul>
	<table class="cassetteitem_other">
		<tbody>
			<tr>
				<td>1</td>
				<td>-</td>
				<td>5階</td>
				<td><span class="cassetteitem_price--rent">12万円</span></td>
				<td><span class="cassetteitem_price--administration">1万円</span></td>
				<td><span class="cassetteitem_price--deposit">2ヶ月</span></td>
				<td><span class="cassetteitem_price--gratuity">1ヶ月</span></td>
				<td><span class="cassetteitem_madori">2LDK</span></td>
				<td><span class="cassetteitem_menseki">50.0m²</span></td>
				<td><a href="/chintai/jnc_000102396495/">詳細を見る</a></td>
			</tr>
		</tbody>
	</table>
</div>
<div class="pagination">
	<a href="?page=1">前へ</a>
	<a href="?page=1">1</a>
	<a href="?page=2">2</a>
</div>
</body>
</html>
`

func TestParseProperties(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(sampleHTML))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	s := NewScraper("https://example.com")
	properties := s.parseProperties(doc)

	if len(properties) != 3 {
		t.Fatalf("Expected 3 properties, got %d", len(properties))
	}

	// Check first property
	p1 := properties[0]
	if p1.Name != "テストマンション" {
		t.Errorf("Property 1 Name = %q, want %q", p1.Name, "テストマンション")
	}
	if p1.Address != "東京都渋谷区渋谷1-1-1" {
		t.Errorf("Property 1 Address = %q, want %q", p1.Address, "東京都渋谷区渋谷1-1-1")
	}
	if p1.Age != 5 {
		t.Errorf("Property 1 Age = %d, want %d", p1.Age, 5)
	}
	if p1.Floor != 3 {
		t.Errorf("Property 1 Floor = %d, want %d", p1.Floor, 3)
	}
	if p1.Rent != 79000 {
		t.Errorf("Property 1 Rent = %f, want %f", p1.Rent, 79000.0)
	}
	if p1.ManagementFee != 5000 {
		t.Errorf("Property 1 ManagementFee = %f, want %f", p1.ManagementFee, 5000.0)
	}
	if p1.Layout != "1K" {
		t.Errorf("Property 1 Layout = %q, want %q", p1.Layout, "1K")
	}
	if p1.Area != 25.5 {
		t.Errorf("Property 1 Area = %f, want %f", p1.Area, 25.5)
	}
	if p1.WalkMinutes != 8 {
		t.Errorf("Property 1 WalkMinutes = %d, want %d", p1.WalkMinutes, 8)
	}
	if p1.ID != "jnc_000102396492" {
		t.Errorf("Property 1 ID = %q, want %q", p1.ID, "jnc_000102396492")
	}

	// Check third property (new construction)
	p3 := properties[2]
	if p3.Age != 0 {
		t.Errorf("Property 3 (new construction) Age = %d, want %d", p3.Age, 0)
	}
	if p3.Rent != 100000 {
		t.Errorf("Property 3 Rent = %f, want %f", p3.Rent, 100000.0)
	}
}

func TestHasNextPage(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name:     "has next page",
			html:     sampleHTML,
			expected: true,
		},
		{
			name:     "no next page",
			html:     sampleHTMLNoNext,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			s := NewScraper("https://example.com")
			got := s.hasNextPage(doc)

			if got != tt.expected {
				t.Errorf("hasNextPage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		page     int
		expected string
	}{
		{
			name:     "URL ending with page=",
			baseURL:  "https://suumo.jp/search?page=",
			page:     5,
			expected: "https://suumo.jp/search?page=5",
		},
		{
			name:     "URL with query params",
			baseURL:  "https://suumo.jp/search?area=tokyo",
			page:     3,
			expected: "https://suumo.jp/search?area=tokyo&page=3",
		},
		{
			name:     "URL without query params",
			baseURL:  "https://suumo.jp/search",
			page:     1,
			expected: "https://suumo.jp/search?page=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScraper(tt.baseURL)
			got := s.buildURL(tt.page)

			if got != tt.expected {
				t.Errorf("buildURL(%d) = %q, want %q", tt.page, got, tt.expected)
			}
		})
	}
}

func TestScraperOptions(t *testing.T) {
	s := NewScraper("https://example.com",
		WithMaxPages(10),
		WithRetryAttempts(5),
		WithRetryDelay(5*time.Second),
	)

	if s.maxPages != 10 {
		t.Errorf("maxPages = %d, want %d", s.maxPages, 10)
	}
	if s.retryAttempts != 5 {
		t.Errorf("retryAttempts = %d, want %d", s.retryAttempts, 5)
	}
	if s.retryDelay != 5*time.Second {
		t.Errorf("retryDelay = %v, want %v", s.retryDelay, 5*time.Second)
	}
}

func TestScrapeWithMockServer(t *testing.T) {
	// Create a mock server
	pageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++
		if pageCount == 1 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(sampleHTML))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(sampleHTMLNoNext))
		}
	}))
	defer server.Close()

	s := NewScraper(server.URL,
		WithMaxPages(5),
		WithRetryAttempts(1),
	)

	ctx := context.Background()
	properties, err := s.Scrape(ctx)
	if err != nil {
		t.Fatalf("Scrape() error = %v", err)
	}

	// Should have 4 unique properties (3 from first page + 1 from second)
	if len(properties) != 4 {
		t.Errorf("Got %d properties, want 4", len(properties))
	}

	// Verify deduplication works (IDs should be unique)
	seen := make(map[string]bool)
	for _, p := range properties {
		if seen[p.ID] {
			t.Errorf("Duplicate ID found: %s", p.ID)
		}
		seen[p.ID] = true
	}
}

func TestScrapeContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleHTML))
	}))
	defer server.Close()

	s := NewScraper(server.URL, WithMaxPages(10))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := s.Scrape(ctx)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
}

func TestFetchPageError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewScraper(server.URL,
		WithRetryAttempts(1),
		WithRetryDelay(1*time.Millisecond),
	)

	ctx := context.Background()
	_, err := s.Scrape(ctx)
	if err == nil {
		t.Error("Expected error for 500 response, got nil")
	}
}
