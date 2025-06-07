package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dlo/twitter-cleanse/internal/cleanse"
)

var (
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
	useCache          bool
	yearsThreshold    float64
	dryRun            bool
)

var rootCmd = &cobra.Command{
	Use:   "twitter-cleanse",
	Short: "Clean up your Twitter follow list",
	Long: `Twitter Cleanse helps you clean up your Twitter follow list by unfollowing users who:
1. Have no tweets, or
2. Haven't tweeted in the last X years (defaults to 2).
3. Have been muted but no longer follow you back.

It will also save these users to distinct Twitter lists for potential re-following.`,
	Run: func(cmd *cobra.Command, args []string) {
		if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessTokenSecret == "" {
			fmt.Println("Error: All authentication credentials are required")
			cmd.Help()
			os.Exit(1)
		}

		config := cleanse.Config{
			ConsumerKey:       consumerKey,
			ConsumerSecret:    consumerSecret,
			AccessToken:       accessToken,
			AccessTokenSecret: accessTokenSecret,
			UseCache:          useCache,
			YearsThreshold:    yearsThreshold,
			DryRun:            dryRun,
		}

		if err := cleanse.Run(config); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&consumerKey, "consumer-key", "", "Your Twitter application's consumer key")
	rootCmd.PersistentFlags().StringVar(&consumerSecret, "consumer-secret", "", "Your Twitter application's consumer secret")
	rootCmd.PersistentFlags().StringVar(&accessToken, "access-token", "", "Your Twitter access token")
	rootCmd.PersistentFlags().StringVar(&accessTokenSecret, "access-token-secret", "", "Your Twitter access token secret")
	rootCmd.PersistentFlags().BoolVar(&useCache, "use-cache", true, "Use a file cache to cache Twitter response payloads")
	rootCmd.PersistentFlags().Float64Var(&yearsThreshold, "years-dormant-threshold", 2.0, "The number of years a person hasn't tweeted for to be unfollowed")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Print out log messages as usual, but don't actually unfollow anyone")

	viper.BindPFlag("consumer-key", rootCmd.PersistentFlags().Lookup("consumer-key"))
	viper.BindPFlag("consumer-secret", rootCmd.PersistentFlags().Lookup("consumer-secret"))
	viper.BindPFlag("access-token", rootCmd.PersistentFlags().Lookup("access-token"))
	viper.BindPFlag("access-token-secret", rootCmd.PersistentFlags().Lookup("access-token-secret"))
}

func initConfig() {
	viper.AutomaticEnv()

	if viper.GetString("consumer-key") != "" {
		consumerKey = viper.GetString("consumer-key")
	}
	if viper.GetString("consumer-secret") != "" {
		consumerSecret = viper.GetString("consumer-secret")
	}
	if viper.GetString("access-token") != "" {
		accessToken = viper.GetString("access-token")
	}
	if viper.GetString("access-token-secret") != "" {
		accessTokenSecret = viper.GetString("access-token-secret")
	}
}
