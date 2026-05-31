package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

const (
	sessionCookieName = "session"
	stateCookieName   = "oauth_state"
	sessionTTL        = 7 * 24 * time.Hour
	stateTTL          = 10 * time.Minute
)

type App struct {
	cfg   Config
	store *Store
	oauth *oauth2.Config
	tmpl  *template.Template
}

func newApp(cfg Config, store *Store, tmpl *template.Template) *App {
	return &App{
		cfg:   cfg,
		store: store,
		oauth: newOAuthConfig(cfg),
		tmpl:  tmpl,
	}
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleHome)
	mux.HandleFunc("/predict", a.handlePredict)
	mux.HandleFunc("/login", a.handleLogin)
	mux.HandleFunc("/callback", a.handleCallback)
	mux.HandleFunc("/logout", a.handleLogout)
	mux.HandleFunc("/login/nationstates", a.handleNSLogin)
	mux.HandleFunc("/leaderboard", a.handleLeaderboard)
	mux.HandleFunc("/games", a.handleGames)
	return mux
}

// currentUser returns the logged-in user, or nil if there is no valid session.
func (a *App) currentUser(r *http.Request) *User {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil
	}
	u, err := a.store.UserBySession(c.Value)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("session lookup error: %v", err)
		}
		return nil
	}
	return u
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	user := a.currentUser(r)
	if err := a.tmpl.ExecuteTemplate(w, "home.html", user); err != nil {
		log.Printf("template error: %v", err)
	}
}

func (a *App) handlePredict(w http.ResponseWriter, r *http.Request) {
	user := a.currentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	switch r.Method {
	case http.MethodGet:
		teams, _ := a.store.AllTeams()
		existing, _ := a.store.UserRankings(user.ID)
		hasSubmitted := len(existing) > 0
		if hasSubmitted {
			sort.Slice(teams, func(i, j int) bool {
				return existing[teams[i].ID] < existing[teams[j].ID]
			})
		}
		a.tmpl.ExecuteTemplate(w, "predict.html", struct {
			Teams        []Team
			HasSubmitted bool
			User         *User
		}{teams, hasSubmitted, user})
	case http.MethodPost:
		if existing, _ := a.store.UserRankings(user.ID); len(existing) > 0 {
			http.Redirect(w, r, "/predict", http.StatusSeeOther)
			return
		}

		ids, err := parseRankingForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := a.store.SaveRankings(user.ID, ids); err != nil {
			http.Error(w, "failed to submit", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/predict", http.StatusSeeOther)
	}
}

func parseRankingForm(r *http.Request) ([]int, error) {
	r.ParseForm()
	raw := r.PostForm["team"]
	if len(raw) != 48 {
		return nil, fmt.Errorf("got %d teams, needed 48", len(raw))
	}

	seen := map[int]bool{}
	ids := make([]int, 0, 48)
	for _, s := range raw {
		id, err := strconv.Atoi(s)
		if err != nil || seen[id] {
			return nil, fmt.Errorf("guh")
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, nil
}

// handleLogin sends the user to Discord, stamping a random state value into a
// cookie so we can verify the callback really originated from this request.
func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randomToken(32)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(stateTTL),
	})
	http.Redirect(w, r, a.oauth.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// handleCallback verifies state, exchanges the code, fetches the Discord
// identity, persists the user, and starts a session.
func (a *App) handleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || r.URL.Query().Get("state") == "" || r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, "invalid OAuth state", http.StatusBadRequest)
		return
	}
	// State consumed — clear it.
	http.SetCookie(w, &http.Cookie{Name: stateCookieName, Path: "/", MaxAge: -1})

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Error(w, "discord authorization denied: "+errParam, http.StatusForbidden)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	token, err := a.oauth.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		log.Printf("token exchange failed: %v", err)
		http.Error(w, "token exchange failed", http.StatusBadGateway)
		return
	}

	du, err := fetchDiscordUser(ctx, a.oauth, token)
	if err != nil {
		log.Printf("fetch discord user failed: %v", err)
		http.Error(w, "could not fetch Discord profile", http.StatusBadGateway)
		return
	}

	user, err := a.store.AddUser("discord", du.ID, du.Username)
	if err != nil {
		log.Printf("upsert user failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := a.startSession(w, user.ID); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	log.Printf("verified discord account: id=%s username=%s (local user id=%d)", du.ID, du.Username, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) startSession(w http.ResponseWriter, userID int64) error {
	tok, err := randomToken(32)
	if err != nil {
		return err
	}
	if err := a.store.CreateSession(userID, tok, sessionTTL); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: tok, Path: "/",
		HttpOnly: true, Secure: a.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode, Expires: time.Now().Add(sessionTTL),
	})
	return nil
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookieName); err == nil {
		if err := a.store.DeleteSession(c.Value); err != nil {
			log.Printf("delete session failed: %v", err)
		}
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	user := a.currentUser(r)
	entries, err := a.store.Leaderboard()
	if err != nil {
		log.Printf("leaderboard query failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	a.tmpl.ExecuteTemplate(w, "leaderboard.html", struct {
		User    *User
		Entries []LeaderboardEntry
	}{user, entries})
}

func (a *App) handleGames(w http.ResponseWriter, r *http.Request) {
	user := a.currentUser(r)
	fixtures, err := a.store.FetchFixtures()
	if err != nil {
		log.Printf("fixtures query failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	a.tmpl.ExecuteTemplate(w, "games.html", struct {
		User     *User
		Fixtures []Fixture
	}{user, fixtures})
}
