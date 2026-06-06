package domain

import "time"

type Difficulty string

const (
	DifficultyEasy     Difficulty = "easy"
	DifficultyMedium   Difficulty = "medium"
	DifficultyHard     Difficulty = "hard"
	DifficultyOlympiad Difficulty = "olympiad"
)

type ProblemType string

const (
	ProblemTypeMultipleChoice ProblemType = "multiple_choice"
	ProblemTypeFillBlank      ProblemType = "fill_blank"
	ProblemTypeSolve          ProblemType = "solve"
	ProblemTypeProof          ProblemType = "proof"
	ProblemTypeOther          ProblemType = "other"
)

type TagCategory string

const (
	TagCategoryTopic  TagCategory = "topic"
	TagCategorySource TagCategory = "source"
	TagCategoryCustom TagCategory = "custom"
)

type ExportFormat string

const (
	ExportFormatLatex ExportFormat = "latex"
	ExportFormatPDF   ExportFormat = "pdf"
)

type ExportVariant string

const (
	ExportVariantStudent ExportVariant = "student"
	ExportVariantAnswer  ExportVariant = "answer"
	ExportVariantBoth    ExportVariant = "both"
)

type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "pending"
	ExportStatusProcessing ExportStatus = "processing"
	ExportStatusDone       ExportStatus = "done"
	ExportStatusFailed     ExportStatus = "failed"
)

type ImportJobStatus string

const (
	ImportJobStatusPending    ImportJobStatus = "pending"
	ImportJobStatusProcessing ImportJobStatus = "processing"
	ImportJobStatusDone       ImportJobStatus = "done"
	ImportJobStatusFailed     ImportJobStatus = "failed"
)

type ImportJob struct {
	ID           string          `json:"id"`
	Filename     string          `json:"filename"`
	InputType    string          `json:"inputType"`
	Status       ImportJobStatus `json:"status"`
	Progress     int             `json:"progress"`
	Result       *ImportPreviewResponse `json:"result,omitempty"`
	ErrorMessage *string         `json:"errorMessage,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	StartedAt    *time.Time      `json:"startedAt,omitempty"`
	CompletedAt  *time.Time      `json:"completedAt,omitempty"`
}

type PaperStatus string

const (
	PaperStatusDraft     PaperStatus = "draft"
	PaperStatusCompleted PaperStatus = "completed"
	PaperStatusReview    PaperStatus = "review"
)

type Tag struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"     validate:"required,max=100"`
	Category     TagCategory `json:"category" validate:"required,oneof=topic source custom"`
	Color        string      `json:"color"    validate:"required"`
	Description  *string     `json:"description,omitempty"`
	ProblemCount int         `json:"problemCount"`
}

type ImageAsset struct {
	ID               string    `json:"id"`
	Filename         string    `json:"filename"`
	MIME             string    `json:"mime"`
	Size             int64     `json:"size"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	URL              string    `json:"url"`
	ThumbnailURL     string    `json:"thumbnailUrl"`
	TagIDs           []string  `json:"tagIds"`
	LinkedProblemIDs []string  `json:"linkedProblemIds"`
	Description      *string   `json:"description,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	IsDeleted        bool      `json:"isDeleted"`
}

type Problem struct {
	ID              string      `json:"id"`
	Code            string      `json:"code"`
	Latex           string      `json:"latex"`
	AnswerLatex     *string     `json:"answerLatex,omitempty"`
	SolutionLatex   *string     `json:"solutionLatex,omitempty"`
	Type            ProblemType `json:"type"`
	Difficulty      Difficulty  `json:"difficulty"`
	SubjectiveScore *float64    `json:"subjectiveScore,omitempty"`
	Subject         *string     `json:"subject,omitempty"`
	Grade           *string     `json:"grade,omitempty"`
	Source          *string     `json:"source,omitempty"`
	TagIDs          []string    `json:"tagIds"`
	ImageIDs        []string    `json:"imageIds"`
	Notes           *string     `json:"notes,omitempty"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
	Version         int         `json:"version"`
	IsDeleted       bool        `json:"isDeleted"`
	Warnings        []string    `json:"warnings,omitempty"`
}

type ProblemDetail struct {
	Problem
	Tags   []Tag        `json:"tags"`
	Images []ImageAsset `json:"images"`
}

type ProblemVersion struct {
	ID        string    `json:"id"`
	ProblemID string    `json:"problemId"`
	Version   int       `json:"version"`
	Snapshot  any       `json:"snapshot"`
	CreatedAt time.Time `json:"createdAt"`
}

type PaperLayout struct {
	Columns           int     `json:"columns"`
	FontSize          int     `json:"fontSize"`
	LineHeight        float64 `json:"lineHeight"`
	PaperSize         string  `json:"paperSize"`
	ShowAnswerVersion bool    `json:"showAnswerVersion"`
}

type PaperItem struct {
	ID            string  `json:"id"`
	ProblemID     string  `json:"problemId"`
	Score         float64 `json:"score"`
	OrderIndex    int     `json:"orderIndex"`
	ImagePosition string  `json:"imagePosition,omitempty"`
	BlankLines    int     `json:"blankLines"`
}

type PaperItemDetail struct {
	PaperItem
	Problem *ProblemDetail `json:"problem,omitempty"`
}

type Paper struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Subtitle   *string     `json:"subtitle,omitempty"`
	SchoolName *string     `json:"schoolName,omitempty"`
	ExamName   *string     `json:"examName,omitempty"`
	Subject    *string     `json:"subject,omitempty"`
	Duration   *int        `json:"duration,omitempty"`
	TotalScore *float64    `json:"totalScore,omitempty"`
	Items      []PaperItem `json:"items"`
	Layout     PaperLayout `json:"layout"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}

type PaperDetail struct {
	Paper
	Description  *string           `json:"description,omitempty"`
	Status       PaperStatus       `json:"status"`
	Instructions *string           `json:"instructions,omitempty"`
	FooterText   *string           `json:"footerText,omitempty"`
	Header       map[string]any    `json:"header"`
	ItemDetails  []PaperItemDetail `json:"itemDetails"`
}

type ExportJob struct {
	ID           string        `json:"id"`
	PaperID      string        `json:"paperId"`
	PaperTitle   string        `json:"paperTitle"`
	Format       ExportFormat  `json:"format"`
	Variant      ExportVariant `json:"variant"`
	Status       ExportStatus  `json:"status"`
	Progress     int           `json:"progress"`
	DownloadURL  *string       `json:"downloadUrl,omitempty"`
	ErrorMessage *string       `json:"errorMessage,omitempty"`
	CreatedAt    time.Time     `json:"createdAt"`
	StartedAt    *time.Time    `json:"startedAt,omitempty"`
	CompletedAt  *time.Time    `json:"completedAt,omitempty"`
}

type SearchCondition struct {
	ID          string  `json:"id"`
	Field       string  `json:"field"`
	Operator    string  `json:"operator"`
	Value       string  `json:"value"`
	SecondValue *string `json:"secondValue,omitempty"`
}

type SearchResult struct {
	ProblemDetail
	Snippet string `json:"snippet"`
}

type SearchHistoryEntry struct {
	ID          string         `json:"id"`
	Query       string         `json:"query"`
	Filters     map[string]any `json:"filters"`
	ResultCount int            `json:"resultCount"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type SavedSearchEntry struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Query     string         `json:"query"`
	Filters   map[string]any `json:"filters"`
	CreatedAt time.Time      `json:"createdAt"`
}

type Paginated[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type MetaStats struct {
	ProblemCount      int `json:"problemCount"`
	ImageCount        int `json:"imageCount"`
	TagCount          int `json:"tagCount"`
	ExportCount       int `json:"exportCount"`
	RecentProblemGain int `json:"recentProblemGain"`
}

type SettingsPayload map[string]any

type UploadedFile struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	Encoding string `json:"encoding,omitempty"`
}

type ImportPreviewRequest struct {
	Latex          string         `json:"latex,omitempty"`
	SeparatorStart string         `json:"separatorStart,omitempty"`
	SeparatorEnd   string         `json:"separatorEnd,omitempty"`
	Defaults       map[string]any `json:"defaults,omitempty"`
	Files          []UploadedFile `json:"files,omitempty"`
}

type ImportPreviewDraft struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Latex         string      `json:"latex"`
	Difficulty    Difficulty  `json:"difficulty"`
	Status        string      `json:"status"`
	Error         *string     `json:"error,omitempty"`
	Warnings      []string    `json:"warnings,omitempty"`
	Subject       *string     `json:"subject,omitempty"`
	Grade         *string     `json:"grade,omitempty"`
	Source        *string     `json:"source,omitempty"`
	TagNames      []string    `json:"tagNames"`
	AnswerLatex   *string     `json:"answerLatex,omitempty"`
	SolutionLatex *string     `json:"solutionLatex,omitempty"`
	InferredType  ProblemType `json:"inferredType,omitempty"`
	NeedsReview   bool        `json:"needsReview,omitempty"`
	SectionTags   []string    `json:"sectionTags,omitempty"`
}

type ImportPreviewResponse struct {
	Parsed            []ImportPreviewDraft `json:"parsed"`
	Errors            []map[string]any     `json:"errors"`
	Warnings          []string             `json:"warnings"`
	PairedAnswerFiles []string             `json:"pairedAnswerFiles,omitempty"`
	UnpairedWarnings  []string             `json:"unpairedWarnings,omitempty"`
}

type ProblemWriteInput struct {
	Latex           string      `json:"latex"           validate:"required,max=50000"`
	AnswerLatex     *string     `json:"answerLatex,omitempty"`
	SolutionLatex   *string     `json:"solutionLatex,omitempty"`
	Type            ProblemType `json:"type"            validate:"required,oneof=multiple_choice fill_blank solve proof other"`
	Difficulty      Difficulty  `json:"difficulty"      validate:"required,oneof=easy medium hard olympiad"`
	SubjectiveScore *float64    `json:"subjectiveScore,omitempty" validate:"omitempty,gte=0,lte=10"`
	Subject         *string     `json:"subject,omitempty"`
	Grade           *string     `json:"grade,omitempty"`
	Source          *string     `json:"source,omitempty"`
	TagIDs          []string    `json:"tagIds"          validate:"dive,required"`
	ImageIDs        []string    `json:"imageIds"        validate:"dive,required"`
	Notes           *string     `json:"notes,omitempty"`
}

type ImageEditInput struct {
	Crop   *CropRect    `json:"crop,omitempty"`
	Rotate *int         `json:"rotate,omitempty"`
	Resize *ResizeInput `json:"resize,omitempty"`
}

type CropRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type ResizeInput struct {
	W int `json:"w"`
	H int `json:"h"`
}

type PaperWriteInput struct {
	Title        string      `json:"title"      validate:"required,max=200"`
	Subtitle     *string     `json:"subtitle,omitempty"`
	SchoolName   *string     `json:"schoolName,omitempty"`
	ExamName     *string     `json:"examName,omitempty"`
	Subject      *string     `json:"subject,omitempty"`
	Duration     *int        `json:"duration,omitempty"`
	TotalScore   *float64    `json:"totalScore,omitempty"`
	Description  *string     `json:"description,omitempty"`
	Status       PaperStatus `json:"status"     validate:"omitempty,oneof=draft completed review"`
	Instructions *string     `json:"instructions,omitempty"`
	FooterText   *string     `json:"footerText,omitempty"`
	Items        []PaperItem `json:"items"`
	Layout       PaperLayout `json:"layout"`
}

type ExportCreateInput struct {
	PaperID string        `json:"paperId" validate:"required"`
	Format  ExportFormat  `json:"format"  validate:"required,oneof=latex pdf"`
	Variant ExportVariant `json:"variant" validate:"required,oneof=student answer both"`
}

type TagWriteInput struct {
	Name        string      `json:"name"     validate:"required,max=100"`
	Category    TagCategory `json:"category" validate:"required,oneof=topic source custom"`
	Color       string      `json:"color"    validate:"required"`
	Description *string     `json:"description,omitempty"`
}

type DemoSeed struct {
	Tags          []SeedTag              `json:"tags"`
	Problems      []SeedProblem          `json:"problems"`
	Images        []SeedImage            `json:"images"`
	Papers        []SeedPaper            `json:"papers"`
	ExportJobs    []SeedExportJob        `json:"exportJobs"`
	SearchHistory []SeedSearchHistory    `json:"searchHistory,omitempty"`
	SavedSearches []SeedSavedSearchEntry `json:"savedSearches,omitempty"`
	Settings      SettingsPayload        `json:"settings"`
}

type SeedTag struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Category    TagCategory `json:"category"`
	Color       string      `json:"color"`
	Description *string     `json:"description,omitempty"`
}

type SeedProblem struct {
	ID              string               `json:"id"`
	Code            string               `json:"code"`
	Latex           string               `json:"latex"`
	AnswerLatex     *string              `json:"answerLatex,omitempty"`
	SolutionLatex   *string              `json:"solutionLatex,omitempty"`
	Type            ProblemType          `json:"type"`
	Difficulty      Difficulty           `json:"difficulty"`
	SubjectiveScore *float64             `json:"subjectiveScore,omitempty"`
	Subject         *string              `json:"subject,omitempty"`
	Grade           *string              `json:"grade,omitempty"`
	Source          *string              `json:"source,omitempty"`
	TagIDs          []string             `json:"tagIds"`
	ImageIDs        []string             `json:"imageIds"`
	Notes           *string              `json:"notes,omitempty"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
	Version         int                  `json:"version"`
	IsDeleted       bool                 `json:"isDeleted"`
	Versions        []SeedProblemVersion `json:"versions,omitempty"`
}

type SeedProblemVersion struct {
	Version   int            `json:"version"`
	CreatedAt time.Time      `json:"createdAt"`
	Snapshot  map[string]any `json:"snapshot"`
}

type SeedImage struct {
	ID               string    `json:"id"`
	Filename         string    `json:"filename"`
	MIME             string    `json:"mime"`
	Size             int64     `json:"size"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	TagIDs           []string  `json:"tagIds"`
	LinkedProblemIDs []string  `json:"linkedProblemIds"`
	Description      *string   `json:"description,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	IsDeleted        bool      `json:"isDeleted"`
}

type SeedPaper struct {
	ID           string      `json:"id"`
	Title        string      `json:"title"`
	Subtitle     *string     `json:"subtitle,omitempty"`
	SchoolName   *string     `json:"schoolName,omitempty"`
	ExamName     *string     `json:"examName,omitempty"`
	Subject      *string     `json:"subject,omitempty"`
	Duration     *int        `json:"duration,omitempty"`
	TotalScore   *float64    `json:"totalScore,omitempty"`
	Description  *string     `json:"description,omitempty"`
	Status       PaperStatus `json:"status"`
	Instructions *string     `json:"instructions,omitempty"`
	FooterText   *string     `json:"footerText,omitempty"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	Layout       PaperLayout `json:"layout"`
	Items        []PaperItem `json:"items"`
}

type SeedExportJob struct {
	ID           string        `json:"id"`
	PaperID      string        `json:"paperId"`
	PaperTitle   string        `json:"paperTitle"`
	Format       ExportFormat  `json:"format"`
	Variant      ExportVariant `json:"variant"`
	Status       ExportStatus  `json:"status"`
	Progress     int           `json:"progress"`
	ErrorMessage *string       `json:"errorMessage,omitempty"`
	CreatedAt    time.Time     `json:"createdAt"`
	CompletedAt  *time.Time    `json:"completedAt,omitempty"`
}

type SeedSearchHistory struct {
	ID          string         `json:"id"`
	Query       string         `json:"query"`
	Filters     map[string]any `json:"filters"`
	ResultCount int            `json:"resultCount"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type SeedSavedSearchEntry struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Query     string         `json:"query"`
	Filters   map[string]any `json:"filters"`
	CreatedAt time.Time      `json:"createdAt"`
}
