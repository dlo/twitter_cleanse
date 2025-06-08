package cleanse

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dlo/twitter-cleanse/internal/cache"
	"github.com/dlo/twitter-cleanse/internal/twitter"
)

type Config struct {
	ClientID       string
	ClientSecret   string
	UseCache       bool
	YearsThreshold float64
	DryRun         bool
}

func Run(config Config) error {
	ctx := context.Background()

	client, err := twitter.NewClient(
		ctx,
		config.ClientID,
		config.ClientSecret,
	)

	var _ *cache.Cache
	if config.UseCache {
		_ = cache.New("twitter_cleanse_cache")
	}

	me, err := client.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	log.Printf("Running cleanse for user: @%s", me.Username)

	following, err := client.GetFollowing(ctx, me.ID)
	if err != nil {
		return fmt.Errorf("failed to get following list: %w", err)
	}

	followers, err := client.GetFollowers(ctx, me.ID)
	if err != nil {
		return fmt.Errorf("failed to get followers list: %w", err)
	}

	followerSet := make(map[string]bool)
	for _, follower := range followers {
		followerSet[follower.ID] = true
	}

	lists, err := client.GetOwnedLists(ctx, me.ID)
	if err != nil {
		return fmt.Errorf("failed to get owned lists: %w", err)
	}

	listMap := make(map[string]string)
	for _, list := range lists {
		listMap[list.Name] = list.ID
	}

	noTweetsListID, err := getOrCreateList(ctx, client, me.ID, listMap, "Unfollowed: No Tweets", "Users who have no tweets.", config.DryRun)
	if err != nil {
		return fmt.Errorf("failed to get/create no tweets list: %w", err)
	}

	quitTwitterListID, err := getOrCreateList(ctx, client, me.ID, listMap, "Unfollowed: Quit Twitter", "Users previously followed who have since stopped tweeting.", config.DryRun)
	if err != nil {
		return fmt.Errorf("failed to get/create quit twitter list: %w", err)
	}

	now := time.Now()
	yearsThreshold := time.Duration(config.YearsThreshold*365*24) * time.Hour

	for _, user := range following {
		if user.Protected {
			log.Printf("Skipping @%s since they are protected", user.Username)
			continue
		}

		if user.MostRecentTweetID == "" {
			log.Printf("Unfollowing @%s since they have no tweets", user.Username)
			if err := unfollowAndAddToList(ctx, client, me.ID, user.ID, noTweetsListID, config.DryRun); err != nil {
				log.Printf("Error unfollowing @%s: %v", user.Username, err)
				continue
			}
		} else {
			tweets, err := client.GetUserTweets(ctx, user.ID, 1)
			if err != nil {
				log.Printf("Error getting tweets for @%s: %v", user.Username, err)
				continue
			}

			if len(tweets) == 0 {
				log.Printf("Unfollowing @%s since they have no accessible tweets", user.Username)
				if err := unfollowAndAddToList(ctx, client, me.ID, user.ID, noTweetsListID, config.DryRun); err != nil {
					log.Printf("Error unfollowing @%s: %v", user.Username, err)
				}
				continue
			}

			lastTweet := tweets[0]
			timeSinceLastTweet := now.Sub(lastTweet.CreatedAt)

			if timeSinceLastTweet > yearsThreshold {
				years := timeSinceLastTweet.Hours() / (365 * 24)
				log.Printf("Unfollowing @%s since they haven't tweeted in %.2f years", user.Username, years)
				if err := unfollowAndAddToList(ctx, client, me.ID, user.ID, quitTwitterListID, config.DryRun); err != nil {
					log.Printf("Error unfollowing @%s: %v", user.Username, err)
				}
			}
		}
	}

	log.Println("Cleanse completed successfully")
	return nil
}

func getOrCreateList(ctx context.Context, client *twitter.Client, userID string, listMap map[string]string, name, description string, dryRun bool) (string, error) {
	if listID, exists := listMap[name]; exists {
		return listID, nil
	}

	if dryRun {
		log.Printf("Would create list: %s", name)
		return "dry-run-list-id", nil
	}

	list, err := client.CreateList(ctx, name, description, true)
	if err != nil {
		return "", err
	}

	return list.ID, nil
}

func unfollowAndAddToList(ctx context.Context, client *twitter.Client, myUserID, targetUserID, listID string, dryRun bool) error {
	if dryRun {
		return nil
	}

	if err := client.AddMemberToList(ctx, listID, targetUserID); err != nil {
		return fmt.Errorf("failed to add to list: %w", err)
	}

	if err := client.UnfollowUser(ctx, myUserID, targetUserID); err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	return nil
}
