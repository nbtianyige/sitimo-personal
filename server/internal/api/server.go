package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mathlib/server/internal/config"
	"mathlib/server/internal/parser"
	"mathlib/server/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type Server struct {
	cfg          config.Config
	svc          *service.Service
	logger       zerolog.Logger
	pdfConverter parser.PDFConverter
}

func New(cfg config.Config, svc *service.Service, logger zerolog.Logger) http.Handler {
	SetLogger(logger)
	server := &Server{
		cfg:          cfg,
		svc:          svc,
		logger:       logger,
		pdfConverter: parser.NewPDFConverter(cfg.PdfConverterType, cfg.MinerUAPIKey, cfg.MinerUAPIBase),
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(server.corsMiddleware)

	r.Get("/healthz", server.handleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/problems", server.handleListProblems)
		r.Get("/problems/{id}", server.handleGetProblem)
		r.Post("/problems", server.handleCreateProblem)
		r.Put("/problems/{id}", server.handleUpdateProblem)
		r.Delete("/problems/{id}", server.handleDeleteProblem)
		r.Post("/problems/{id}/restore", server.handleRestoreProblem)
		r.Delete("/problems/{id}/hard", server.handleHardDeleteProblem)
		r.Get("/problems/{id}/versions", server.handleProblemVersions)
		r.Post("/problems/{id}/versions/{version}", server.handleRollbackProblemVersion)
		r.Post("/problems/batch-import/preview", server.handlePreviewImport)
		r.Post("/problems/batch-import", server.handleCommitImport)
		r.Post("/problems/batch-tag", server.handleBatchTagProblems)
		r.Post("/problems/batch-delete", server.handleBatchDeleteProblems)

		r.Get("/images", server.handleListImages)
		r.Get("/images/{id}", server.handleGetImage)
		r.Post("/images", server.handleUploadImage)
		r.Put("/images/{id}", server.handleUpdateImage)
		r.Delete("/images/{id}", server.handleDeleteImage)
		r.Delete("/images/{id}/hard", server.handleHardDeleteImage)
		r.Post("/images/{id}/restore", server.handleRestoreImage)
		r.Post("/images/{id}/edit", server.handleEditImage)
		r.Get("/images/{id}/file", server.handleImageFile)
		r.Get("/images/{id}/thumbnail", server.handleImageThumbnail)
		r.Post("/images/batch-delete", server.handleBatchDeleteImages)

		r.Get("/tags", server.handleListTags)
		r.Post("/tags", server.handleCreateTag)
		r.Put("/tags/{id}", server.handleUpdateTag)
		r.Delete("/tags/{id}", server.handleDeleteTag)
		r.Post("/tags/{id}/merge", server.handleMergeTag)

		r.Get("/search", server.handleSearch)
		r.Get("/search/history", server.handleSearchHistory)
		r.Delete("/search/history/{id}", server.handleDeleteSearchHistory)
		r.Get("/search/saved", server.handleSavedSearches)
		r.Post("/search/saved", server.handleCreateSavedSearch)
		r.Delete("/search/saved/{id}", server.handleDeleteSavedSearch)

		r.Get("/papers", server.handleListPapers)
		r.Get("/papers/{id}", server.handleGetPaper)
		r.Post("/papers", server.handleCreatePaper)
		r.Put("/papers/{id}", server.handleUpdatePaper)
		r.Put("/papers/{id}/items", server.handleUpdatePaperItems)
		r.Delete("/papers/{id}", server.handleDeletePaper)
		r.Post("/papers/{id}/duplicate", server.handleDuplicatePaper)

		r.Post("/exports", server.handleCreateExport)
		r.Get("/exports", server.handleListExports)
		r.Get("/exports/{id}", server.handleGetExport)
		r.Delete("/exports/{id}", server.handleDeleteExport)
		r.Get("/exports/{id}/download", server.handleDownloadExport)
		r.Get("/exports/stream", server.handleExportStream)

		r.Post("/imports", server.handleCreateImport)
		r.Get("/imports", server.handleListImports)
		r.Get("/imports/{id}", server.handleGetImport)
		r.Delete("/imports/{id}", server.handleDeleteImport)

		r.Get("/meta/grades", server.handleMetaGrades)
		r.Get("/meta/stats", server.handleMetaStats)
		r.Get("/meta/recent-problems", server.handleRecentProblems)
		r.Get("/meta/recent-exports", server.handleRecentExports)

		r.Get("/settings", server.handleGetSettings)
		r.Put("/settings", server.handleUpdateSettings)
		r.Get("/settings/demo-data/status", server.handleDemoDataStatus)

		r.Route("/settings", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(server.adminAuth)
				r.Post("/demo-data", server.handleResetDemoData)
				r.Post("/demo-data/load", server.handleLoadDemoData)
				r.Post("/demo-data/clear", server.handleClearDemoData)
				r.Post("/export-all", server.handleExportAll)
				r.Post("/import-all", server.handleImportAll)
				r.Post("/sweep-orphans", server.handleSweepOrphans)
			})
		})
	})

	return r
}

func (s *Server) adminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Admin-Key")
		if key == "" || key != s.cfg.AdminAPIKey() {
			respondError(w, http.StatusUnauthorized, "unauthorized", errors.New("admin key required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err == nil && cookie.Value != "" {
		return true
	}
	auth := r.Header.Get("Authorization")
	return auth != ""
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Repository().Ping(r.Context()); err != nil {
		respondError(w, http.StatusServiceUnavailable, "db_unavailable", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	allowed := strings.Join(s.cfg.AllowedOrigins, ", ")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && containsOrigin(s.cfg.AllowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		} else if origin == "" && allowed != "" {
			w.Header().Set("Access-Control-Allow-Origin", s.cfg.AllowedOrigins[0])
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func containsOrigin(origins []string, origin string) bool {
	for _, item := range origins {
		if item == origin {
			return true
		}
	}
	return false
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseBoolPtr(raw string) *bool {
	if raw == "" {
		return nil
	}
	switch raw {
	case "true", "1", "yes":
		value := true
		return &value
	case "false", "0", "no":
		value := false
		return &value
	default:
		return nil
	}
}

func parseFloatPtr(raw string) *float64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &value
}

func handleNotFound(w http.ResponseWriter, err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		respondError(w, http.StatusNotFound, "not_found", errors.New("资源不存在"))
		return true
	}
	return false
}

func parseCommaList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

func streamJSON(ctx context.Context, w http.ResponseWriter, payload any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := w.Write([]byte("data: "))
		if err != nil {
			return err
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := w.Write(raw); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n\n")); err != nil {
			return err
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		return nil
	}
}

func requestContext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 30*time.Second)
}
