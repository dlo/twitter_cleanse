package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func getToken(url, clientID, clientSecret string) string {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Printf("status %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body)
}

func main() {
	// cmd.Execute()
	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok {
		panic("API_KEY is not set")
	}

	apiSecret, ok := os.LookupEnv("API_SECRET")
	if !ok {
		panic("API_SECRET is not set")
	}
	token := getToken("https://api.x.com/2/oauth2/token", apiKey, apiSecret)
	fmt.Println(token)

	// req, err := http.NewRequest("GET", "https://api.twitter.com/2/users/by/username/dwlz", nil)

	// if err != nil {
	// 	panic(err)
	// }
	// req.Header.Set("Authorization", "Bearer "+os.Getenv("BEARER_TOKEN"))

	// res, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	panic(err)
	// }
	// body, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(string(body))
}
