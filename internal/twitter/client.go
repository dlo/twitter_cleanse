package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
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

func NewClient(ctx context.Context, clientID, clientSecret string) (*Client, error) {
	httpClient, err := GetXHTTPClient(ctx, clientID, clientSecret)
	if err != nil {
		return nil, err
	}
	return &Client{httpClient: httpClient}, nil
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

		endpoint := fmt.Sprintf("https://api.x.com/2/users/%s/following", userID)
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
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
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response T
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

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
	// https://api.twitter.com/2/users/{source_user_id}/following/{target_user_id}
	_, err := deleteRequest[any](ctx, c, fmt.Sprintf("https://api.x.com/2/users/%s/following/%s", sourceUserID, targetUserID))
	if err != nil {
		return err
	}

	return nil
}
