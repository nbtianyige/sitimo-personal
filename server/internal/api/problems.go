package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/service"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListProblems(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	params := service.ProblemListParams{
		Keyword:    r.URL.Query().Get("keyword"),
		Subject:    r.URL.Query().Get("subject"),
		Grade:      r.URL.Query().Get("grade"),
		Difficulty: parseCommaList(r.URL.Query().Get("difficulty")),
		Type:       r.URL.Query().Get("type"),
		TagIDs:     parseCommaList(r.URL.Query().Get("tagIds")),
		HasImage:   parseBoolPtr(r.URL.Query().Get("hasImage")),
		ScoreMin:   parseFloatPtr(r.URL.Query().Get("scoreMin")),
		ScoreMax:   parseFloatPtr(r.URL.Query().Get("scoreMax")),
		SortBy:     r.URL.Query().Get("sortBy"),
		SortOrder:  r.URL.Query().Get("sortOrder"),
		Page:       parseInt(r.URL.Query().Get("page"), 1),
		PageSize:   parseInt(r.URL.Query().Get("pageSize"), 24),
		Deleted:    r.URL.Query().Get("deleted") == "true",
	}

	result, err := s.svc.ListProblems(ctx, params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_problems_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetProblem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	item, err := s.svc.Repository().GetProblemDetail(ctx, chi.URLParam(r, "id"), true)
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "get_problem_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handleCreateProblem(w http.ResponseWriter, r *http.Request) {
	var input domain.ProblemWriteInput
	if !decodeAndValidate(w, r, &input) {
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()

	item, warnings, err := s.svc.CreateProblem(ctx, input)
	if err != nil {
		respondError(w, http.StatusBadRequest, "create_problem_failed", err)
		return
	}
	respondJSON(w, http.StatusCreated, map[string]any{"problem": item, "warnings": warnings})
}

func (s *Server) handleUpdateProblem(w http.ResponseWriter, r *http.Request) {
	var input domain.ProblemWriteInput
	if !decodeAndValidate(w, r, &input) {
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()

	item, warnings, err := s.svc.UpdateProblem(ctx, chi.URLParam(r, "id"), input)
	if handleNotFound(w, err) {
		return
	}
	if err != nil {
		respondError(w, http.StatusBadRequest, "update_problem_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"problem": item, "warnings": warnings})
}

func (s *Server) handleDeleteProblem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().SoftDeleteProblem(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "delete_problem_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleRestoreProblem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().RestoreProblem(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "restore_problem_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleHardDeleteProblem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().HardDeleteProblem(ctx, chi.URLParam(r, "id")); err != nil {
		respondError(w, http.StatusInternalServerError, "hard_delete_problem_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleProblemVersions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.Repository().ListProblemVersions(ctx, chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list_versions_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleRollbackProblemVersion(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	version, err := strconv.Atoi(chi.URLParam(r, "version"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_version", errors.New("版本号无效"))
		return
	}
	item, err := s.svc.Repository().RollbackProblemVersion(ctx, chi.URLParam(r, "id"), version)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "rollback_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (s *Server) handlePreviewImport(w http.ResponseWriter, r *http.Request) {
	var input domain.ImportPreviewRequest

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(5 << 20); err != nil {
			respondError(w, http.StatusBadRequest, "parse_multipart_failed", err)
			return
		}

		defaultsStr := r.FormValue("defaults")
		if defaultsStr != "" {
			if err := json.Unmarshal([]byte(defaultsStr), &input.Defaults); err != nil {
				respondError(w, http.StatusBadRequest, "invalid_defaults_json", err)
				return
			}
		}
		if input.Defaults == nil {
			input.Defaults = make(map[string]any)
		}

		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			respondError(w, http.StatusBadRequest, "no_files", errors.New("请至少上传一个 .tex 或 .md 文件"))
			return
		}

		for _, fh := range files {
			ext := strings.ToLower(filepath.Ext(fh.Filename))
			if ext != ".tex" && ext != ".md" && ext != ".pdf" {
				respondError(w, http.StatusBadRequest, "invalid_file_type",
					errors.New(fmt.Sprintf("仅支持 .tex、.md 或 .pdf 文件: %s", fh.Filename)))
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

			// If PDF, try to convert via external tool to .tex / .md
			if ext == ".pdf" {
				converted, convExt, err := s.pdfConverter.Convert(content, fh.Filename)
				if err != nil {
					respondError(w, http.StatusBadRequest, "pdf_conversion_failed", err)
					return
				}
				content = []byte(converted)
				// Change the effective extension for downstream parser routing
				fh.Filename = strings.TrimSuffix(fh.Filename, ".pdf") + convExt
			}

			input.Files = append(input.Files, domain.UploadedFile{
				Filename: fh.Filename,
				Content:  content,
			})
		}
	} else {
		if err := decodeJSON(r, &input); err != nil {
			respondError(w, http.StatusBadRequest, "invalid_json", err)
			return
		}
	}

	respondJSON(w, http.StatusOK, s.svc.PreviewBatchImport(input))
}

func (s *Server) handleCommitImport(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Drafts []domain.ImportPreviewDraft `json:"drafts"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := s.svc.CommitBatchImport(ctx, payload.Drafts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "import_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (s *Server) handleBatchTagProblems(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProblemIDs []string `json:"problemIds"`
		TagIDs     []string `json:"tagIds"`
		Replace    bool     `json:"replace"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().BatchTagProblems(ctx, payload.ProblemIDs, payload.TagIDs, payload.Replace); err != nil {
		respondError(w, http.StatusInternalServerError, "batch_tag_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleBatchDeleteProblems(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProblemIDs []string `json:"problemIds"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := s.svc.Repository().BatchDeleteProblems(ctx, payload.ProblemIDs); err != nil {
		respondError(w, http.StatusInternalServerError, "batch_delete_failed", err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
