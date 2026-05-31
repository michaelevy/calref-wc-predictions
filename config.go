package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type Config struct {
	Port           string
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	DatabasePath   string
	CookieSecure   bool
	NSSiteToken    string
	NSUserAgent    string
	APIFootballKey string
	FIFAUserAgent  string
}

func loadConfig() Config {
	loadEnvFile(".env")

	cfg := Config{
		Port:           getenv("PORT", "8080"),
		ClientID:       os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret:   os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURL:    getenv("DISCORD_REDIRECT_URL", "http://localhost:8080/callback"),
		DatabasePath:   getenv("DATABASE_PATH", "app.db"),
		CookieSecure:   os.Getenv("COOKIE_SECURE") == "true",
		NSSiteToken:    os.Getenv("NS_SITE_TOKEN"),
		NSUserAgent:    "calref-wc-predictions by Catiania (michaelfeehanlevy@gmail.com)",
		FIFAUserAgent:  "calref-wc-predictions by Michael (michaelfeehanlevy@gmail.com)",
		APIFootballKey: os.Getenv("API_FOOTBALL_KEY"),
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		log.Fatal("DISCORD_CLIENT_ID and DISCORD_CLIENT_SECRET must be set (copy .env.example to .env and fill them in)")
	}
	return cfg
}

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // no .env file — rely on real env vars
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
	sc.Err()
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
