package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// InitWebHandlers - sets up the http handlers
func InitWebHandlers(r *httprouter.Router) {
	r.GET("/", limit(handleIndex, ratelimiter))
	r.GET("/docs", limit(handleDocs, ratelimiter))
	r.Handler(http.MethodGet, "/assets/*filepath", http.StripPrefix("/assets",
		http.FileServer(http.Dir("./assets"))))
}

// handleIndex handles index
func handleIndex(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	templates.ExecuteTemplate(res, "index.html", nil)
}

// handleIndex handles index
func handleDocs(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	templates.ExecuteTemplate(res, "docs.html", nil)
}
