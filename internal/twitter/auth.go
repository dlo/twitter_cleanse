package twitter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	pkce "github.com/nirasan/go-oauth-pkce-code-verifier"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

const (
	authURL  = "https://x.com/i/oauth2/authorize"
	tokenURL = "https://api.x.com/2/oauth2/token"
	redirect = "https://twitter.lionheartsw.com/callback" // must be allowed in your X app
)

type TokenSourceFunc func() (*oauth2.Token, error)

func (f TokenSourceFunc) Token() (*oauth2.Token, error) {
	return f()
}

func GetXHTTPClient(ctx context.Context, clientID, clientSecret string, forceRefresh bool) (*http.Client, error) {
	cfgDir := func() string { d, _ := os.UserConfigDir(); return filepath.Join(d, "x-oauth-demo") }()
	tokenPath := filepath.Join(cfgDir, "token.json")

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret, // leave empty for pure-PKCE public client
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirect,
		Scopes: []string{
			"tweet.read",
			"users.read",
			"list.read",
			"follows.read",
			"offline.access",
		},
	}

	// helpers for cache
	load := func() (*oauth2.Token, error) {
		f, err := os.Open(tokenPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var t oauth2.Token
		return &t, json.NewDecoder(f).Decode(&t)
	}
	save := func(t *oauth2.Token) {
		_ = os.MkdirAll(cfgDir, 0o700)
		f, err := os.Create(tokenPath)
		if err != nil {
			log.Fatalf("save token: %v", err)
		}
		defer f.Close()
		_ = json.NewEncoder(f).Encode(t)
	}

	// ── 1. Obtain initial token (cache ⟹ browser + paste) ────────────────────
	var tok *oauth2.Token
	if !forceRefresh {
		tok, _ = load()
		if tok != nil && tok.Valid() {
			log.Printf("DEBUG: Loaded cached token - Access: %s, Refresh: %s, Expires: %v",
				tok.AccessToken,
				tok.RefreshToken,
				tok.Expiry)
			// ── 2. Wrap in a persist-on-refresh TokenSource ──────────────────────────
			ts := oauth2.ReuseTokenSource(tok, TokenSourceFunc(func() (*oauth2.Token, error) {
				ctx := context.Background()
				newTok, err := conf.TokenSource(ctx, tok).Token()
				if err == nil {
					log.Printf("DEBUG: Refreshed token - Access: %s, Refresh: %s, Expires: %v",
						newTok.AccessToken,
						newTok.RefreshToken,
						newTok.Expiry)
					save(newTok)
				}
				return newTok, err
			}))

			return oauth2.NewClient(ctx, ts), nil
		}
	}

	// Token is invalid or missing, start OAuth flow
	verifier, _ := pkce.CreateCodeVerifier()
	challenge := verifier.CodeChallengeS256()
	state := uuid.NewString()

	consentURL := conf.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	fmt.Println("\n*** ACTION REQUIRED ***")
	fmt.Printf("1) Your browser will open the X authorisation page.\n")
	fmt.Printf("2) Authorise the app.\n")
	fmt.Printf("3) When X redirects to %s … copy the ENTIRE URL\n", redirect)
	fmt.Printf("4) Paste it here and press <Enter>.\n\n")

	_ = browser.OpenURL(consentURL)

	fmt.Print("Paste redirect URL: ")
	input := bufio.NewScanner(os.Stdin)
	if !input.Scan() {
		return nil, fmt.Errorf("no input")
	}

	rawURL := strings.TrimSpace(input.Text())
	if err := input.Err(); err != nil {
		return nil, err
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	if q.Get("state") != state {
		return nil, fmt.Errorf("state mismatch")
	}

	code := q.Get("code")
	if code == "" {
		return nil, fmt.Errorf("code not found in URL")
	}

	tok, err = conf.Exchange(
		ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier.String()),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	log.Printf("DEBUG: New token exchanged - Access: %s..., Refresh: %s..., Expires: %v",
		tok.AccessToken[:min(20, len(tok.AccessToken))],
		tok.RefreshToken[:min(20, len(tok.RefreshToken))],
		tok.Expiry)

	save(tok)

	// ── 2. Wrap in a persist-on-refresh TokenSource ──────────────────────────
	ts := oauth2.ReuseTokenSource(tok, TokenSourceFunc(func() (*oauth2.Token, error) {
		newTok, err := conf.TokenSource(ctx, tok).Token()
		if err == nil {
			save(newTok)
		}
		return newTok, err
	}))

	return oauth2.NewClient(ctx, ts), nil
}
