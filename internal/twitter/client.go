package twitter

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	httpClient        *http.Client
}

type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Name          string `json:"name"`
	PublicMetrics struct {
		FollowersCount int `json:"followers_count"`
		FollowingCount int `json:"following_count"`
		TweetCount     int `json:"tweet_count"`
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
	ID          string `json:"id"`
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

func NewClient(consumerKey, consumerSecret, accessToken, accessTokenSecret string) *Client {
	return &Client{
		ConsumerKey:       consumerKey,
		ConsumerSecret:    consumerSecret,
		AccessToken:       accessToken,
		AccessTokenSecret: accessTokenSecret,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (c *Client) generateSignature(method, requestURL string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramStrs []string
	for _, k := range keys {
		paramStrs = append(paramStrs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramString := strings.Join(paramStrs, "&")

	baseString := fmt.Sprintf("%s&%s&%s",
		method,
		requestURL,
		url.QueryEscape(paramString))

	signingKey := fmt.Sprintf("%s&%s", c.ConsumerSecret, c.AccessTokenSecret)

	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(baseString))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c *Client) buildAuthHeader(method, requestURL string, params map[string]string) string {
	nonce := c.generateNonce()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	oauthParams := map[string]string{
		"oauth_consumer_key":     c.ConsumerKey,
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        timestamp,
		"oauth_token":            c.AccessToken,
		"oauth_version":          "1.0",
	}

	allParams := make(map[string]string)
	for k, v := range oauthParams {
		allParams[k] = v
	}
	for k, v := range params {
		allParams[k] = v
	}

	signature := c.generateSignature(method, requestURL, allParams)
	oauthParams["oauth_signature"] = signature

	var authParts []string
	for k, v := range oauthParams {
		authParts = append(authParts, fmt.Sprintf(`%s="%s"`, k, url.QueryEscape(v)))
	}

	return "OAuth " + strings.Join(authParts, ", ")
}

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, params map[string]string) (*http.Response, error) {
	baseURL := "https://api.twitter.com/2"
	requestURL := baseURL + endpoint

	var req *http.Request
	var err error

	if method == "GET" {
		if len(params) > 0 {
			u, _ := url.Parse(requestURL)
			q := u.Query()
			for k, v := range params {
				q.Set(k, v)
			}
			u.RawQuery = q.Encode()
			requestURL = u.String()
		}
		req, err = http.NewRequestWithContext(ctx, method, requestURL, nil)
	} else {
		data := url.Values{}
		for k, v := range params {
			data.Set(k, v)
		}
		req, err = http.NewRequestWithContext(ctx, method, requestURL, strings.NewReader(data.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		return nil, err
	}

	authHeader := c.buildAuthHeader(method, strings.Split(requestURL, "?")[0], params)
	req.Header.Set("Authorization", authHeader)

	return c.httpClient.Do(req)
}

func (c *Client) GetFollowing(ctx context.Context, userID string) ([]User, error) {
	var allUsers []User
	nextToken := ""

	for {
		params := map[string]string{
			"user.fields": "public_metrics",
			"max_results": "1000",
		}
		if nextToken != "" {
			params["pagination_token"] = nextToken
		}

		endpoint := fmt.Sprintf("/users/%s/following", userID)
		resp, err := c.makeRequest(ctx, "GET", endpoint, params)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
		}

		var response UsersResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		allUsers = append(allUsers, response.Data...)

		if response.Meta.NextToken == "" {
			break
		}
		nextToken = response.Meta.NextToken
	}

	return allUsers, nil
}

func (c *Client) GetFollowers(ctx context.Context, userID string) ([]User, error) {
	var allUsers []User
	nextToken := ""

	for {
		params := map[string]string{
			"user.fields": "public_metrics",
			"max_results": "1000",
		}
		if nextToken != "" {
			params["pagination_token"] = nextToken
		}

		endpoint := fmt.Sprintf("/users/%s/followers", userID)
		resp, err := c.makeRequest(ctx, "GET", endpoint, params)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
		}

		var response UsersResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		allUsers = append(allUsers, response.Data...)

		if response.Meta.NextToken == "" {
			break
		}
		nextToken = response.Meta.NextToken
	}

	return allUsers, nil
}

func (c *Client) GetUserTweets(ctx context.Context, userID string, maxResults int) ([]Tweet, error) {
	params := map[string]string{
		"tweet.fields": "created_at",
		"max_results":  strconv.Itoa(maxResults),
	}

	endpoint := fmt.Sprintf("/users/%s/tweets", userID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []Tweet{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response TweetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *Client) UnfollowUser(ctx context.Context, sourceUserID, targetUserID string) error {
	endpoint := fmt.Sprintf("/users/%s/following/%s", sourceUserID, targetUserID)
	resp, err := c.makeRequest(ctx, "DELETE", endpoint, map[string]string{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) GetMe(ctx context.Context) (*User, error) {
	params := map[string]string{
		"user.fields": "public_metrics",
	}

	resp, err := c.makeRequest(ctx, "GET", "/users/me", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *Client) GetOwnedLists(ctx context.Context, userID string) ([]List, error) {
	endpoint := fmt.Sprintf("/users/%s/owned_lists", userID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, map[string]string{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response ListsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *Client) CreateList(ctx context.Context, name, description string, private bool) (*List, error) {
	params := map[string]string{
		"name":        name,
		"description": description,
		"private":     strconv.FormatBool(private),
	}

	resp, err := c.makeRequest(ctx, "POST", "/lists", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data List `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func (c *Client) AddMemberToList(ctx context.Context, listID, userID string) error {
	endpoint := fmt.Sprintf("/lists/%s/members", listID)
	params := map[string]string{
		"user_id": userID,
	}

	resp, err := c.makeRequest(ctx, "POST", endpoint, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
