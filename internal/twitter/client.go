package twitter

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	httpClient   *http.Client
	cacheDir     string
	cacheTTL     time.Duration
	cacheEnabled bool
}

type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Name          string `json:"name"`
	Protected     bool   `json:"protected"`
	PublicMetrics struct {
		TweetCount int `json:"tweet_count"`
	} `json:"public_metrics"`
}

type Tweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	AuthorID  string    `json:"author_id"`
}

type UsersResponse struct {
	Data []User `json:"data"`
	Meta struct {
		NextToken   string `json:"next_token,omitempty"`
		ResultCount int    `json:"result_count"`
	} `json:"meta"`
}

type TweetsResponse struct {
	Data []Tweet `json:"data"`
	Meta struct {
		NextToken   string `json:"next_token,omitempty"`
		ResultCount int    `json:"result_count"`
	} `json:"meta"`
}

type List struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

type ListsResponse struct {
	Data []List `json:"data"`
	Meta struct {
		ResultCount int `json:"result_count"`
	} `json:"meta"`
}

type cacheEntry struct {
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

func getCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (c *Client) getCachePath(key string) string {
	return filepath.Join(c.cacheDir, key+".json")
}

func (c *Client) readFromCache(key string) (json.RawMessage, bool) {
	if !c.cacheEnabled {
		log.Printf("[CACHE] Cache disabled, skipping read for key: %s", key[:8]+"...")
		return nil, false
	}

	cachePath := c.getCachePath(key)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		log.Printf("[CACHE] Cache miss - file not found for key: %s (%s)", key[:8]+"...", err.Error())
		return nil, false
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		log.Printf("[CACHE] Cache miss - invalid cache entry for key: %s (%s)", key[:8]+"...", err.Error())
		return nil, false
	}

	// Check if cache entry is expired
	age := time.Since(entry.Timestamp)
	if age > c.cacheTTL {
		log.Printf("[CACHE] Cache miss - expired entry (age: %v, TTL: %v) for key: %s", age, c.cacheTTL, key[:8]+"...")
		os.Remove(cachePath) // Clean up expired cache
		return nil, false
	}

	log.Printf("[CACHE] Cache hit - entry age: %v, TTL: %v, key: %s", age, c.cacheTTL, key[:8]+"...")
	return entry.Data, true
}

func (c *Client) writeToCache(key string, data json.RawMessage) {
	if !c.cacheEnabled {
		log.Printf("[CACHE] Cache disabled, skipping write for key: %s", key[:8]+"...")
		return
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		log.Printf("[CACHE] Failed to create cache directory: %s (%s)", c.cacheDir, err.Error())
		return
	}

	entry := cacheEntry{
		Data:      data,
		Timestamp: time.Now(),
	}

	entryData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("[CACHE] Failed to marshal cache entry for key: %s (%s)", key[:8]+"...", err.Error())
		return
	}

	cachePath := c.getCachePath(key)
	if err := os.WriteFile(cachePath, entryData, 0644); err != nil {
		log.Printf("[CACHE] Failed to write cache file for key: %s (%s)", key[:8]+"...", err.Error())
		return
	}

	log.Printf("[CACHE] Successfully cached response (size: %d bytes) for key: %s", len(data), key[:8]+"...")
}

// ClearCache removes all cached files
func (c *Client) ClearCache() error {
	if !c.cacheEnabled {
		log.Printf("[CACHE] Cache disabled, skipping clear operation")
		return nil
	}

	log.Printf("[CACHE] Clearing all cache files from directory: %s", c.cacheDir)
	if err := os.RemoveAll(c.cacheDir); err != nil {
		log.Printf("[CACHE] Failed to clear cache directory: %s (%s)", c.cacheDir, err.Error())
		return err
	}

	log.Printf("[CACHE] Successfully cleared all cache files")
	return nil
}

func NewClient(ctx context.Context, clientID, clientSecret string) (*Client, error) {
	httpClient, err := GetXHTTPClient(ctx, clientID, clientSecret, false)
	if err != nil {
		return nil, err
	}

	// Set up default cache configuration
	cacheDir := "./cache"
	cacheTTL := 15 * time.Minute

	log.Printf("[CACHE] Initializing Twitter client with cache enabled - Dir: %s, TTL: %v", cacheDir, cacheTTL)

	return &Client{
		httpClient:   httpClient,
		cacheDir:     cacheDir,
		cacheTTL:     cacheTTL,
		cacheEnabled: true,
	}, nil
}

func deleteRequest[T any](ctx context.Context, client *Client, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response T
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func postRequest[T any](ctx context.Context, client *Client, url string, body any) (*T, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response T
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func getRequest[T any](ctx context.Context, client *Client, url string) (*T, error) {
	log.Printf("[HTTP] GET request to: %s", url)

	// Try to get from cache first
	cacheKey := getCacheKey(url)
	log.Printf("[CACHE] Checking cache for URL: %s (key: %s)", url, cacheKey[:8]+"...")

	if cachedData, found := client.readFromCache(cacheKey); found {
		var response T
		if err := json.Unmarshal(cachedData, &response); err == nil {
			log.Printf("[HTTP] Returning cached response for: %s", url)
			return &response, nil
		}
		log.Printf("[CACHE] Failed to unmarshal cached data for key: %s, making HTTP request", cacheKey[:8]+"...")
	}

	// Make HTTP request
	log.Printf("[HTTP] Making live API request to: %s", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("[HTTP] Failed to create request: %s", err.Error())
		return nil, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		log.Printf("[HTTP] Request failed: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[HTTP] Response status: %d for URL: %s", resp.StatusCode, url)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[HTTP] API error %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response T
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[HTTP] Failed to read response body: %s", err.Error())
		return nil, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("[HTTP] Failed to unmarshal response: %s", err.Error())
		return nil, err
	}

	log.Printf("[HTTP] Successfully parsed response (size: %d bytes) for: %s", len(body), url)

	// Cache the successful response
	client.writeToCache(cacheKey, json.RawMessage(body))

	return &response, nil
}

func (c *Client) GetMe(ctx context.Context) (*User, error) {
	response, err := getRequest[struct {
		Data User `json:"data"`
	}](ctx, c, "https://api.x.com/2/users/me")
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *Client) GetOwnedLists(ctx context.Context, userID string) ([]List, error) {
	response, err := getRequest[ListsResponse](ctx, c, fmt.Sprintf("https://api.x.com/2/users/%s/owned_lists", userID))
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *Client) GetFollowers(ctx context.Context, userID string) ([]User, error) {
	response, err := getRequest[UsersResponse](ctx, c, fmt.Sprintf("https://api.x.com/2/users/%s/followers", userID))
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *Client) GetUserTweets(ctx context.Context, userID string, maxResults int) ([]Tweet, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) CreateList(ctx context.Context, name, description string, private bool) (*List, error) {
	type ListResponse struct {
		Data List `json:"data"`
	}
	response, err := postRequest[ListResponse](ctx, c, "https://api.x.com/2/lists", List{
		Name:        name,
		Description: description,
		Private:     private,
	})
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *Client) AddMemberToList(ctx context.Context, listID, userID string) error {
	_, err := postRequest[any](ctx, c, fmt.Sprintf("https://api.x.com/2/lists/%s/members", listID), map[string]string{"user_id": userID})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UnfollowUser(ctx context.Context, sourceUserID, targetUserID string) error {
	_, err := deleteRequest[any](ctx, c, fmt.Sprintf("https://api.x.com/2/users/%s/following/%s", sourceUserID, targetUserID))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetFollowing(ctx context.Context, userID string) ([]User, error) {
	var allUsers []User

	var paginationToken string
	for {
		url := fmt.Sprintf("https://api.x.com/2/users/%s/following", userID)
		if paginationToken != "" {
			url = fmt.Sprintf("%s&pagination_token=%s", url, paginationToken)
		}
		response, err := getRequest[UsersResponse](ctx, c, url)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, response.Data...)

		if response.Meta.NextToken == "" {
			break
		}

		paginationToken = response.Meta.NextToken
	}

	return allUsers, nil
}
