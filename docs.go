package main

import (
	"embed"
	"net/http"
)

//go:embed openapi.yaml
var specFS embed.FS

const docsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>nano-dataset-manager API</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css"/>
<style>html { box-sizing: border-box; } *, *:before, *:after { box-sizing: inherit; } body { margin: 0; }</style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger-ui" })</script>
</body>
</html>`

func (s *Server) registerDocs(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.yaml", s.handleSpec)
	mux.HandleFunc("GET /docs", s.handleDocs)
}

func (s *Server) handleSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	http.ServeFileFS(w, r, specFS, "openapi.yaml")
}

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(docsHTML))
}
