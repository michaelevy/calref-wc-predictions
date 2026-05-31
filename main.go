package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/*.html
var templatesFS embed.FS

func main() {
	cfg := loadConfig()

	store, err := openStore(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	app := newApp(cfg, store, tmpl)

	addr := ":" + cfg.Port
	log.Printf("listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, app.routes()); err != nil {
		log.Fatal(err)
	}
}
