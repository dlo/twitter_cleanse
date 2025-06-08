package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/dlo/twitter-cleanse/internal/cleanse"
)

var (
	clientID       = flag.String("client-id", getEnv("CLIENT_ID", ""), "Your Twitter application's client ID")
	clientSecret   = flag.String("client-secret", getEnv("CLIENT_SECRET", ""), "Your Twitter application's client secret")
	useCache       = flag.Bool("use-cache", true, "Use a file cache to cache Twitter response payloads")
	yearsThreshold = flag.Float64("years-dormant-threshold", 2.0, "The number of years a person hasn't tweeted for to be unfollowed")
	dryRun         = flag.Bool("dry-run", false, "Print out log messages as usual, but don't actually unfollow anyone")
	help           = flag.Bool("help", false, "Show this help message")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func showUsage() {
	fmt.Fprintf(os.Stderr, `Twitter Cleanse - Clean up your Twitter follow list

Unfollows users who:
1. Have no tweets
2. Haven't tweeted in X years (default: 2)
3. Are muted and don't follow back

Users are saved to Twitter lists for later reference.

Usage: %s [OPTIONS]

Options:
`, os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n  CLIENT_ID      Twitter application's client ID\n  CLIENT_SECRET  Twitter application's client secret\n")
}

func Execute() {
	flag.Usage = showUsage
	flag.Parse()

	if *help {
		showUsage()
		os.Exit(0)
	}

	if *clientID == "" || *clientSecret == "" {
		fmt.Fprintf(os.Stderr, "Error: All authentication credentials are required\n\n")
		showUsage()
		os.Exit(1)
	}

	config := cleanse.Config{
		ClientID:       *clientID,
		ClientSecret:   *clientSecret,
		UseCache:       *useCache,
		YearsThreshold: *yearsThreshold,
		DryRun:         *dryRun,
	}

	if err := cleanse.Run(config); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
