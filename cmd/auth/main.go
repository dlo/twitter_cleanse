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

	twitterClient, err := twitter.NewClient(context.Background(), clientID, clientSecret)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	me, err := twitterClient.GetMe(ctx)
	if err != nil {
		log.Fatal(err)
	}

	following, err := twitterClient.GetFollowing(ctx, me.ID)
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range following {
		fmt.Println(user.Username)
	}
}
