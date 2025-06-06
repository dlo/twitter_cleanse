# Twitter Cleanse (Go)

![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A Go utility to clean up your Twitter follow list by unfollowing inactive users. This is a complete rewrite of the original Python version, now using the Twitter v2 API.

## What it does

This utility unfollows everyone you follow who:

1. Has no tweets, or
2. Hasn't tweeted in the last X years (defaults to 2).

It will also save these users to distinct Twitter lists in case you'd like to re-follow them in the future:
- **"Unfollowed: No Tweets"** - Users who have never tweeted
- **"Unfollowed: Quit Twitter"** - Users who haven't tweeted in X years

## Installation

### From Source

```bash
git clone https://github.com/dlo/twitter-cleanse.git
cd twitter-cleanse
go build -o twitter-cleanse .
```

### Using Go Install

```bash
go install github.com/dlo/twitter-cleanse@latest
```

## Setup

Before you start, you'll need to create a [Twitter Developer Account](https://developer.twitter.com/) and create an app to get your API credentials:

1. Go to [Twitter Developer Portal](https://developer.twitter.com/en/portal/dashboard)
2. Create a new app
3. Generate your API keys and tokens:
   - Consumer Key (API Key)
   - Consumer Secret (API Secret)
   - Access Token
   - Access Token Secret

Make sure your app has **Read and Write** permissions to manage follows and create lists.

## Usage

### Command Line Arguments

```bash
twitter-cleanse \
  --consumer-key "your_consumer_key" \
  --consumer-secret "your_consumer_secret" \
  --access-token "your_access_token" \
  --access-token-secret "your_access_token_secret"
```

### Environment Variables

You can also set credentials via environment variables:

```bash
export CONSUMER_KEY="your_consumer_key"
export CONSUMER_SECRET="your_consumer_secret"
export ACCESS_TOKEN="your_access_token"
export ACCESS_TOKEN_SECRET="your_access_token_secret"

twitter-cleanse
```

### Options

- `--consumer-key` - Your Twitter application's consumer key
- `--consumer-secret` - Your Twitter application's consumer secret  
- `--access-token` - Your Twitter access token
- `--access-token-secret` - Your Twitter access token secret
- `--use-cache` - Use file cache for API responses (default: true)
- `--years-dormant-threshold` - Years of inactivity before unfollowing (default: 2.0)
- `--dry-run` - Show what would be done without actually unfollowing (default: false)

### Examples

**Dry run to see what would happen:**
```bash
twitter-cleanse --dry-run --consumer-key "..." --consumer-secret "..." --access-token "..." --access-token-secret "..."
```

**Unfollow users inactive for 1 year:**
```bash
twitter-cleanse --years-dormant-threshold 1.0 --consumer-key "..." --consumer-secret "..." --access-token "..." --access-token-secret "..."
```

**Disable caching:**
```bash
twitter-cleanse --use-cache=false --consumer-key "..." --consumer-secret "..." --access-token "..." --access-token-secret "..."
```

## How it works

1. **Authentication**: Uses OAuth 1.0a to authenticate with Twitter v2 API
2. **Data Collection**: Fetches your following list and their tweet counts
3. **Analysis**: For each user you follow:
   - If they have 0 tweets → adds to "No Tweets" list and unfollows
   - If their last tweet is older than threshold → adds to "Quit Twitter" list and unfollows
4. **List Management**: Creates private lists to store unfollowed users for potential re-following
5. **Caching**: Optionally caches API responses to avoid rate limits on repeated runs

## Rate Limiting

The Twitter v2 API has rate limits. This tool:
- Respects rate limits and will pause when necessary
- Uses caching by default to minimize API calls
- Processes users sequentially to avoid overwhelming the API

## Migration from Python Version

This Go version provides the same functionality as the original Python version but with:
- Better performance and lower memory usage
- Native OAuth 1.0a implementation (no external auth libraries)
- Improved error handling and logging
- Support for environment variables
- Built-in dry-run mode

## Contributing

Feel free to send a pull request! If your addition fixes a bug or adds a useful feature, it'll most likely make it in.

## License

Apache License, Version 2.0. See [LICENSE](LICENSE) file for details.

## Differences from Original

This Go version uses the Twitter v2 API instead of v1.1, which means:
- More reliable and modern API endpoints
- Better rate limiting handling
- Improved data structures and response formats
- Future-proof as Twitter phases out v1.1 API

