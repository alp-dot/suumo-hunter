// Package scraper provides functionality to scrape property listings from SUUMO.
package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/avast/retry-go/v4"

	"github.com/alp/suumo-hunter-go/internal/models"
)

const (
	// DefaultMaxPages is the default maximum number of pages to scrape.
	DefaultMaxPages = 30

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultRetryAttempts is the default number of retry attempts.
	DefaultRetryAttempts = 3

	// DefaultRetryDelay is the default delay between retries.
	DefaultRetryDelay = 10 * time.Second

	// UserAgent is the User-Agent header sent with requests.
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// Scraper fetches property listings from SUUMO.
type Scraper struct {
	client        *http.Client
	maxPages      int
	retryAttempts uint
	retryDelay    time.Duration
	baseURL       string
}

// Option is a function that configures a Scraper.
type Option func(*Scraper)

// WithMaxPages sets the maximum number of pages to scrape.
func WithMaxPages(n int) Option {
	return func(s *Scraper) {
		s.maxPages = n
	}
}

// WithRetryAttempts sets the number of retry attempts.
func WithRetryAttempts(n uint) Option {
	return func(s *Scraper) {
		s.retryAttempts = n
	}
}

// WithRetryDelay sets the delay between retries.
func WithRetryDelay(d time.Duration) Option {
	return func(s *Scraper) {
		s.retryDelay = d
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(s *Scraper) {
		s.client = c
	}
}

// NewScraper creates a new Scraper with the given options.
func NewScraper(baseURL string, opts ...Option) *Scraper {
	s := &Scraper{
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		maxPages:      DefaultMaxPages,
		retryAttempts: DefaultRetryAttempts,
		retryDelay:    DefaultRetryDelay,
		baseURL:       baseURL,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Scrape fetches all property listings from SUUMO.
// It paginates through the search results up to maxPages.
func (s *Scraper) Scrape(ctx context.Context) ([]models.Property, error) {
	var allProperties []models.Property
	seenIDs := make(map[string]bool)

	for page := 1; page <= s.maxPages; page++ {
		select {
		case <-ctx.Done():
			return allProperties, ctx.Err()
		default:
		}

		properties, hasMore, err := s.scrapePage(ctx, page)
		if err != nil {
			return allProperties, fmt.Errorf("failed to scrape page %d: %w", page, err)
		}

		// Deduplicate properties
		for _, p := range properties {
			if !seenIDs[p.ID] {
				seenIDs[p.ID] = true
				allProperties = append(allProperties, p)
			}
		}

		if !hasMore {
			break
		}
	}

	return allProperties, nil
}

// scrapePage fetches a single page of property listings.
// Returns the properties found and whether there are more pages.
func (s *Scraper) scrapePage(ctx context.Context, page int) ([]models.Property, bool, error) {
	url := s.buildURL(page)

	var doc *goquery.Document
	err := retry.Do(
		func() error {
			var fetchErr error
			doc, fetchErr = s.fetchPage(ctx, url)
			return fetchErr
		},
		retry.Attempts(s.retryAttempts),
		retry.Delay(s.retryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			// Log retry attempt (in production, use proper logging)
			fmt.Printf("Retry %d for %s: %v\n", n+1, url, err)
		}),
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch page after retries: %w", err)
	}

	properties := s.parseProperties(doc)
	hasMore := s.hasNextPage(doc)

	return properties, hasMore, nil
}

// buildURL constructs the URL for a specific page.
func (s *Scraper) buildURL(page int) string {
	// The base URL should end with "page=" or have a page parameter
	if strings.Contains(s.baseURL, "page=") {
		// Replace existing page parameter
		return s.baseURL + strconv.Itoa(page)
	}
	// Append page parameter
	if strings.Contains(s.baseURL, "?") {
		return s.baseURL + "&page=" + strconv.Itoa(page)
	}
	return s.baseURL + "?page=" + strconv.Itoa(page)
}

// fetchPage fetches and parses a single page.
func (s *Scraper) fetchPage(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}

// parseProperties extracts all properties from a page.
func (s *Scraper) parseProperties(doc *goquery.Document) []models.Property {
	var properties []models.Property

	// Each property listing is in a div.cassetteitem
	doc.Find("div.cassetteitem").Each(func(_ int, item *goquery.Selection) {
		props := s.parsePropertyItem(item)
		properties = append(properties, props...)
	})

	return properties
}

// parsePropertyItem extracts property information from a single listing.
// A single listing can contain multiple rooms/units.
func (s *Scraper) parsePropertyItem(item *goquery.Selection) []models.Property {
	var properties []models.Property

	// Common information for all rooms in this listing
	name := strings.TrimSpace(item.Find("div.cassetteitem_content-title").Text())
	address := strings.TrimSpace(item.Find("li.cassetteitem_detail-col1").Text())

	// Parse building age and floors from detail-col3
	var buildingAge, buildingFloors string
	item.Find("li.cassetteitem_detail-col3 div").Each(func(i int, div *goquery.Selection) {
		text := strings.TrimSpace(div.Text())
		if i == 0 {
			buildingAge = text
		} else if i == 1 {
			buildingFloors = text
		}
	})

	age, _ := models.ParseAge(buildingAge)

	// Parse access information (walking time to station)
	// Take the first station's walking time
	var walkMinutes int
	item.Find("li.cassetteitem_detail-col2 div.cassetteitem_detail-text").Each(func(i int, div *goquery.Selection) {
		if i == 0 {
			text := strings.TrimSpace(div.Text())
			walkMinutes, _ = models.ParseWalkMinutes(text)
		}
	})

	// Each room/unit is in a table row
	item.Find("table.cassetteitem_other tbody tr").Each(func(_ int, row *goquery.Selection) {
		prop := s.parseRoomRow(row, name, address, age, walkMinutes, buildingFloors)
		if prop.ID != "" {
			properties = append(properties, prop)
		}
	})

	return properties
}

// parseRoomRow extracts information for a single room/unit.
func (s *Scraper) parseRoomRow(row *goquery.Selection, name, address string, age, walkMinutes int, buildingFloors string) models.Property {
	// Floor
	floorText := strings.TrimSpace(row.Find("td").Eq(2).Text())
	floor, _ := models.ParseFloor(floorText)

	// If floor parsing fails, try to get from building floors
	if floor == 0 && buildingFloors != "" {
		floor, _ = models.ParseFloor(buildingFloors)
	}

	// Rent
	rentText := strings.TrimSpace(row.Find("span.cassetteitem_price--rent").Text())
	rent, _ := models.ParseRent(rentText)

	// Management fee
	managementFeeText := strings.TrimSpace(row.Find("span.cassetteitem_price--administration").Text())
	managementFee, _ := models.ParseRent(managementFeeText)

	// Deposit
	deposit := strings.TrimSpace(row.Find("span.cassetteitem_price--deposit").Text())

	// Key money
	keyMoney := strings.TrimSpace(row.Find("span.cassetteitem_price--gratuity").Text())

	// Layout
	layout := strings.TrimSpace(row.Find("span.cassetteitem_madori").Text())

	// Area
	areaText := strings.TrimSpace(row.Find("span.cassetteitem_menseki").Text())
	area, _ := models.ParseArea(areaText)

	// URL and ID
	var url, id string
	row.Find("td a").Each(func(_ int, a *goquery.Selection) {
		if href, exists := a.Attr("href"); exists {
			if strings.Contains(href, "/chintai/") {
				url = "https://suumo.jp" + href
				id = models.ExtractPropertyID(href)
			}
		}
	})

	return models.Property{
		ID:            id,
		Name:          name,
		Address:       address,
		Age:           age,
		Floor:         floor,
		Rent:          rent,
		ManagementFee: managementFee,
		Deposit:       deposit,
		KeyMoney:      keyMoney,
		Layout:        layout,
		Area:          area,
		WalkMinutes:   walkMinutes,
		URL:           url,
	}
}

// hasNextPage checks if there are more pages to scrape.
func (s *Scraper) hasNextPage(doc *goquery.Document) bool {
	// Check for pagination - look for "次へ" (next) link
	hasNext := false
	doc.Find("div.pagination a").Each(func(_ int, a *goquery.Selection) {
		if strings.Contains(a.Text(), "次へ") {
			hasNext = true
		}
	})
	return hasNext
}
