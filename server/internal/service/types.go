package service

import "mathlib/server/internal/domain"

type ProblemListParams struct {
	Keyword    string
	Subject    string
	Grade      string
	Difficulty []string
	Type       string
	TagIDs     []string
	HasImage   *bool
	ScoreMin   *float64
	ScoreMax   *float64
	SortBy     string
	SortOrder  string
	Page       int
	PageSize   int
	Deleted    bool
}

type ImageListParams struct {
	Keyword  string
	TagIDs   []string
	MIME     string
	Page     int
	PageSize int
	Deleted  bool
}

type PaperListParams struct {
	Keyword  string
	Page     int
	PageSize int
}

type ExportListParams struct {
	Status   string
	Page     int
	PageSize int
}

type ImportListParams struct {
	Status   string
	Page     int
	PageSize int
}

type ImportJobCreateInput struct {
	Filename  string
	InputType string
	Files     []domain.UploadedFile
	Defaults  map[string]any
}

type SeedEnvelope struct {
	Tags          []domain.SeedTag              `json:"tags"`
	Problems      []domain.SeedProblem          `json:"problems"`
	Images        []domain.SeedImage            `json:"images"`
	Papers        []domain.SeedPaper            `json:"papers"`
	ExportJobs    []domain.SeedExportJob        `json:"exportJobs"`
	SearchHistory []domain.SeedSearchHistory    `json:"searchHistory,omitempty"`
	SavedSearches []domain.SeedSavedSearchEntry `json:"savedSearches,omitempty"`
	Settings      domain.SettingsPayload        `json:"settings"`
}

func (s SeedEnvelope) ToDomain() domain.DemoSeed {
	return domain.DemoSeed{
		Tags:          s.Tags,
		Problems:      s.Problems,
		Images:        s.Images,
		Papers:        s.Papers,
		ExportJobs:    s.ExportJobs,
		SearchHistory: s.SearchHistory,
		SavedSearches: s.SavedSearches,
		Settings:      s.Settings,
	}
}
