package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"log"
	"net/http"
	"x-proxy/app"
	"x-proxy/cmd/proxy"
)

func main() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.Application.WebPort),
		Handler: Routes(),
	}

	app.Application.Log.Info(fmt.Sprintf("Starting server on :%s", app.Application.WebPort))
	err := server.ListenAndServe()

	if err != nil {
		log.Fatalln(err)
	}
}

func Routes() http.Handler {

	router := chi.NewRouter()

	// specify who is allowed to connect
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	
	router.HandleFunc("/*", proxy.Proxy)
	return router
}
