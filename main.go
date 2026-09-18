package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"internet-manager-go/internal/config"
	"internet-manager-go/internal/handlers"
	"internet-manager-go/internal/mikrotik"
	"internet-manager-go/internal/models"
)

//go:embed templates/* static/*
var content embed.FS

func main() {
	cfg := config.Load()

	tmpl, err := template.ParseFS(content, "templates/*.html")
	if err != nil {
		log.Fatal("templates:", err)
	}

	mt := mikrotik.New(cfg.APIHost, cfg.APIUser, cfg.APIPass, cfg.APIPort)
	store := models.NewStore(cfg.SettingsFile)
	app := handlers.New(cfg, mt, store, tmpl)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(content)))
	app.Register(mux)

	log.Printf("Internet Manager on :%s → API %s:%s\n", cfg.WebPort, cfg.APIHost, cfg.APIPort)
	log.Fatal(http.ListenAndServe(":"+cfg.WebPort, mux))
}
