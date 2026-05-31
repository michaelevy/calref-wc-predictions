package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const nsVerifyAPI = "https://www.nationstates.net/cgi-bin/api.cgi"

func canonNation(name string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "_")
}

func verifyNation(ctx context.Context, userAgent, nation, code, siteToken string) (bool, error) {
	q := url.Values{}
	q.Set("a", "verify")
	q.Set("nation", canonNation(nation))
	q.Set("checksum", strings.TrimSpace(code))
	if siteToken != "" {
		q.Set("token", siteToken)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nsVerifyAPI+"?"+q.Encode(), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return false, fmt.Errorf("ns verify HTTP %d: %s", resp.StatusCode, body)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 16))
	return strings.TrimSpace(string(body)) == "1", nil
}

func (a *App) nsVerifyLoginURL() string {
	if a.cfg.NSSiteToken != "" {
		return "https://www.nationstates.net/page=verify_login?token=" + url.QueryEscape(a.cfg.NSSiteToken)
	}
	return "https://www.nationstates.net/page=verify_login"
}

func (a *App) handleNSLogin(w http.ResponseWriter, r *http.Request) {
	render := func(errMsg string) {
		a.tmpl.ExecuteTemplate(w, "nsLogin.html", map[string]any{
			"VerifyURL": a.nsVerifyLoginURL(),
			"Error":     errMsg,
		})
	}

	switch r.Method {
	case http.MethodGet:
		render("")
	case http.MethodPost:
		r.ParseForm()
		nation := r.FormValue("nation")
		code := r.FormValue("code")
		if nation == "" || code == "" {
			render("Enter both your nation name and the code.")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		ok, err := verifyNation(ctx, a.cfg.NSUserAgent, nation, code, a.cfg.NSSiteToken)
		if err != nil {
			log.Printf("ns verify error: %v", err)
			http.Error(w, "could not reach NS", http.StatusBadGateway)
			return
		}
		if !ok {
			render("Verification failed")
			return
		}
		id := canonNation(nation)
		user, err := a.store.AddUser("nationstates", id, id)
		if err != nil {
			log.Printf("failed to add NS user: %v", err)
			http.Error(w, "failed to add NS user", http.StatusInternalServerError)
			return
		}
		if err := a.startSession(w, user.ID); err != nil {
			log.Printf("failed to create session for NS user: %v", err)
			http.Error(w, "failed to create NS session", http.StatusInternalServerError)
			return
		}
		log.Printf("verified nation: %s %d", id, user.ID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
