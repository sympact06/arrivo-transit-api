package handlers

import (
	"net/http"
	"path/filepath"

	httpSwagger "github.com/swaggo/http-swagger"
)

// SwaggerHandler serves the Swagger UI and OpenAPI specification
type SwaggerHandler struct {
	apiSpecPath string
}

// NewSwaggerHandler creates a new Swagger handler
func NewSwaggerHandler(apiSpecPath string) *SwaggerHandler {
	return &SwaggerHandler{
		apiSpecPath: apiSpecPath,
	}
}

// ServeSwaggerUI serves the Swagger UI interface
func (h *SwaggerHandler) ServeSwaggerUI() http.HandlerFunc {
	return httpSwagger.Handler(
		httpSwagger.URL("/api/v1/swagger/doc.json"), // The url pointing to API definition
	)
}

// ServeOpenAPISpec serves the OpenAPI specification JSON/YAML
func (h *SwaggerHandler) ServeOpenAPISpec() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Determine content type based on file extension or Accept header
		contentType := "application/json"
		if filepath.Ext(h.apiSpecPath) == ".yaml" || filepath.Ext(h.apiSpecPath) == ".yml" {
			contentType = "application/x-yaml"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		http.ServeFile(w, r, h.apiSpecPath)
	}
}

// ServeStaticDocs serves static documentation files
func (h *SwaggerHandler) ServeStaticDocs(docsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security: prevent directory traversal
		cleanPath := filepath.Clean(r.URL.Path)
		if cleanPath == "/" {
			cleanPath = "/index.html"
		}

		filePath := filepath.Join(docsDir, cleanPath)
		
		// Ensure the file is within the docs directory
		if !isSubPath(docsDir, filePath) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		http.ServeFile(w, r, filePath)
	}
}

// isSubPath checks if child is a subdirectory of parent
func isSubPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return !filepath.IsAbs(rel) && !filepath.HasPrefix(rel, "..")
}