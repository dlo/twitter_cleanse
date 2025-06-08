package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dlo/twitter-cleanse/internal/twitter"
)

func main() {
	clientID := os.Getenv("OAUTH_2_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_2_CLIENT_SECRET")
	client, err := twitter.GetXHTTPClient(context.Background(), clientID, clientSecret)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(client)
}
