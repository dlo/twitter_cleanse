package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
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
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	return &Client{
		httpClient: config.Client(oauth1.NoContext, token),
	}
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

// func (c *Client) GetFollowers(ctx context.Context, userID string) ([]User, error) {
// 	var allUsers []User
// 	nextToken := ""

// 	for {
// 		params := map[string]string{
// 			"user.fields": "public_metrics",
// 			"max_results": "1000",
// 		}
// 		if nextToken != "" {
// 			params["pagination_token"] = nextToken
// 		}

// 		endpoint := fmt.Sprintf("/users/%s/followers", userID)
// 		resp, err := c.makeRequest(ctx, "GET", endpoint, params)
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer resp.Body.Close()

// 		if resp.StatusCode != http.StatusOK {
// 			body, _ := io.ReadAll(resp.Body)
// 			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
// 		}

// 		var response UsersResponse
// 		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 			return nil, err
// 		}

// 		allUsers = append(allUsers, response.Data...)

// 		if response.Meta.NextToken == "" {
// 			break
// 		}
// 		nextToken = response.Meta.NextToken
// 	}

// 	return allUsers, nil
// }

// func (c *Client) UnfollowUser(ctx context.Context, sourceUserID, targetUserID string) error {
// 	endpoint := fmt.Sprintf("/users/%s/following/%s", sourceUserID, targetUserID)
// 	resp, err := c.makeRequest(ctx, "DELETE", endpoint, map[string]string{})
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
// 	}

// 	return nil
// }

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

// func (c *Client) GetUserTweets(ctx context.Context, userID string, maxResults int) ([]Tweet, error) {
// 	params := map[string]string{
// 		"tweet.fields": "created_at",
// 		"max_results":  strconv.Itoa(maxResults),
// 	}

// 	endpoint := fmt.Sprintf("/users/%s/tweets", userID)
// 	resp, err := c.makeRequest(ctx, "GET", endpoint, params)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusNotFound {
// 		return []Tweet{}, nil
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
// 	}

// 	var response TweetsResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 		return nil, err
// 	}

// 	return response.Data, nil
// }

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

// func (c *Client) CreateList(ctx context.Context, name, description string, private bool) (*List, error) {
// 	params := map[string]string{
// 		"name":        name,
// 		"description": description,
// 		"private":     strconv.FormatBool(private),
// 	}

// 	resp, err := c.makeRequest(ctx, "POST", "/lists", params)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusCreated {
// 		body, _ := io.ReadAll(resp.Body)
// 		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
// 	}

// 	var response struct {
// 		Data List `json:"data"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 		return nil, err
// 	}

// 	return &response.Data, nil
// }

// func (c *Client) AddMemberToList(ctx context.Context, listID, userID string) error {
// 	endpoint := fmt.Sprintf("/lists/%s/members", listID)
// 	params := map[string]string{
// 		"user_id": userID,
// 	}

// 	resp, err := c.makeRequest(ctx, "POST", endpoint, params)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
// 	}

// 	return nil
// }
