package main

import (
	"context"
	"embed"
	"html/template"
	"log"
	"net/http"
	"time"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed templates/*.html resources/*
var assetsFS embed.FS

func main() {
	cfg := loadConfig()

	store, err := openStore(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	app := newApp(cfg, store, tmpl, assetsFS)

	go func() {
		refresh := func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			resp, err := app.fetchFixtures(ctx)
			if err != nil {
				log.Printf("fetch fixtures: %v", err)
				return
			}
			if err := store.RefreshFixtures(resp); err != nil {
				log.Printf("refresh fixtures: %v", err)
			}
		}
		refresh()
		for range time.NewTicker(30 * time.Minute).C {
			refresh()
		}
	}()

	addr := ":" + cfg.Port
	log.Printf("listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, app.routes()); err != nil {
		log.Fatal(err)
	}
}
