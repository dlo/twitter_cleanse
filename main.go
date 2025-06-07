package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dlo/twitter-cleanse/internal/twitter"
)

func main() {
	ctx := context.Background()
	t := twitter.NewClient(os.Getenv("API_KEY"), os.Getenv("API_SECRET"), os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))

	me, err := t.GetMe(ctx)
	if err != nil {
		panic(err)
	}

	lists, err := t.GetOwnedLists(ctx, me.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println(lists)
}
