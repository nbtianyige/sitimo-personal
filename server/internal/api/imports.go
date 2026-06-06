package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/service"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleCreateImport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "parse_multipart_failed", err)
		return
	}

	var defaults map[string]any
	defaultsStr := r.FormValue("defaults")
	if defaultsStr != "" {
		if err := json.Unmarshal([]byte(defaultsStr), &defaults); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_defaults_json", err)
			return
		}
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		respondError(w, http.StatusBadRequest, "no_files", errors.New("请至少上传一个 .tex、.md 或 .pdf 文件"))
		return
	}

	var uploaded []domain.UploadedFile
	for _, fh := range files {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if ext != ".tex" && ext != ".md" && ext != ".pdf" {
			respondError(w, http.StatusBadRequest, "invalid_file_type",
				errors.New("仅支持 .tex、.md 或 .pdf 文件: "+fh.Filename))
			return
		}

		file, err := fh.Open()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "file_read_failed", err)
			return
		}
		content, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "file_read_failed", err)
			return
		}

		if ext == ".pdf" {
			converted, convExt, err := s.pdfConverter.Convert(content, fh.Filename)
			if err != nil {
				respondError(w, http.StatusBadRequest, "pdf_conversion_failed", err)
				return
			}
			content = []byte(converted)
			fh.Filename = strings.TrimSuffix(fh.Filename, ".pdf") + convExt
		}

		uploaded = append(uploaded, domain.UploadedFile{
			Filename: fh.Filename,
			Content:  content,
		})
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	item, err := s.svc.CreateImportJob(ctx, service.ImportJobCreateInput{
		Filename:  files[0].Filename,
		InputType: "file_upload",
		Files:     uploaded,
		Defaults:  defaults,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "create_import_failed", err)
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

func (s *Server) handleListImports(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.ListImportJobs(ctx, service.ImportListParams{
		Status:   r.URL.Query().Get("status"),
		Page:     parseInt(r.URL.Query().Get("page"), 1),
		PageSize: parseInt(r.URL.Query().Get("pageSize"), 24),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_imports_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleGetImport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	item, err := s.svc.GetImportJob(ctx, chi.URLParam(r, "id"))
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "get_import_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleDeleteImport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.DeleteImportJob(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "delete_import_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
