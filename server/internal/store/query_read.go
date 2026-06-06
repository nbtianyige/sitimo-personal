package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/search"
	"mathlib/server/internal/store/sqlc"

	"github.com/jackc/pgx/v5"
)

type ProblemListOptions struct {
	Keyword        string
	Subject        string
	Grade          string
	Difficulty     []string
	Type           string
	TagIDs         []string
	HasImage       *bool
	ScoreMin       *float64
	ScoreMax       *float64
	SortBy         string
	SortOrder      string
	Page           int
	PageSize       int
	IncludeDeleted bool
}

type SearchOptions struct {
	Keyword    string
	Formula    string
	Conditions []domain.SearchCondition
}

type ImageListOptions struct {
	Keyword        string
	TagIDs         []string
	MIME           string
	Page           int
	PageSize       int
	IncludeDeleted bool
}

type PaperListOptions struct {
	Keyword        string
	Page           int
	PageSize       int
	IncludeDeleted bool
}

type ExportListOptions struct {
	Status   string
	Page     int
	PageSize int
}

type ImportListOptions struct {
	Status   string
	Page     int
	PageSize int
}

type ImageDetail struct {
	Image          domain.ImageAsset      `json:"image"`
	LinkedProblems []domain.ProblemDetail `json:"linkedProblems"`
	Tags           []domain.Tag           `json:"tags"`
}

type exportNotificationPayload struct {
	JobID    string `json:"jobId"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

type sqlArgs struct {
	values []any
}

func (a *sqlArgs) Add(value any) string {
	a.values = append(a.values, value)
	return fmt.Sprintf("$%d", len(a.values))
}

type rowScanner interface {
	Scan(dest ...any) error
}

type problemPredicateOptions struct {
	IncludeDeleted bool
	Keyword        string
	Formula        string
	Subject        string
	Grade          string
	Difficulty     []string
	Type           string
	TagIDs         []string
	HasImage       *bool
	ScoreMin       *float64
	ScoreMax       *float64
	Conditions     []domain.SearchCondition
}

type problemPredicateState struct {
	clause     string
	args       []any
	keywordArg string
}

type searchRow struct {
	Problem sqlc.Problem
	Snippet string
}

func problemRecordFromGetProblemByIDRow(record sqlc.GetProblemByIDRow) sqlc.Problem {
	return sqlc.Problem{
		ID:              record.ID,
		Code:            record.Code,
		Latex:           record.Latex,
		AnswerLatex:     record.AnswerLatex,
		SolutionLatex:   record.SolutionLatex,
		ProblemType:     record.ProblemType,
		Difficulty:      record.Difficulty,
		SubjectiveScore: record.SubjectiveScore,
		Subject:         record.Subject,
		Grade:           record.Grade,
		Source:          record.Source,
		Notes:           record.Notes,
		SearchTsv:       record.SearchTsv,
		FormulaTokens:   record.FormulaTokens,
		FormulaTsv:      record.FormulaTsv,
		Version:         record.Version,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
		DeletedAt:       record.DeletedAt,
	}
}

func problemRecordFromGetProblemsByIDsRow(record sqlc.GetProblemsByIDsRow) sqlc.Problem {
	return sqlc.Problem{
		ID:              record.ID,
		Code:            record.Code,
		Latex:           record.Latex,
		AnswerLatex:     record.AnswerLatex,
		SolutionLatex:   record.SolutionLatex,
		ProblemType:     record.ProblemType,
		Difficulty:      record.Difficulty,
		SubjectiveScore: record.SubjectiveScore,
		Subject:         record.Subject,
		Grade:           record.Grade,
		Source:          record.Source,
		Notes:           record.Notes,
		SearchTsv:       record.SearchTsv,
		FormulaTokens:   record.FormulaTokens,
		FormulaTsv:      record.FormulaTsv,
		Version:         record.Version,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
		DeletedAt:       record.DeletedAt,
	}
}

func (r *Repository) GetProblemDetail(ctx context.Context, id string, includeDeleted bool) (*domain.ProblemDetail, error) {
	record, err := r.queries.GetProblemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !includeDeleted && record.DeletedAt.Valid {
		return nil, pgx.ErrNoRows
	}

	items, err := r.buildProblemDetails(ctx, []sqlc.Problem{problemRecordFromGetProblemByIDRow(record)}, includeDeleted)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}

	return &items[0], nil
}

func (r *Repository) GetImage(ctx context.Context, id string, includeDeleted bool) (*domain.ImageAsset, error) {
	records, err := r.loadImageAssetsByIDs(ctx, []string{id}, includeDeleted)
	if err != nil {
		return nil, err
	}
	item, ok := records[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return &item, nil
}

func (r *Repository) GetImageDetail(ctx context.Context, id string, includeDeleted bool) (*ImageDetail, error) {
	imageAsset, err := r.GetImage(ctx, id, includeDeleted)
	if err != nil {
		return nil, err
	}

	tagMap, err := r.loadTagMapByIDs(ctx, imageAsset.TagIDs)
	if err != nil {
		return nil, err
	}
	linkedProblems, err := r.loadProblemDetailsByIDs(ctx, imageAsset.LinkedProblemIDs, includeDeleted)
	if err != nil {
		return nil, err
	}

	detail := &ImageDetail{
		Image:          *imageAsset,
		LinkedProblems: linkedProblems,
		Tags:           make([]domain.Tag, 0, len(imageAsset.TagIDs)),
	}
	for _, tagID := range imageAsset.TagIDs {
		if tag, ok := tagMap[tagID]; ok {
			detail.Tags = append(detail.Tags, tag)
		}
	}

	return detail, nil
}

func (r *Repository) GetPaperDetail(ctx context.Context, id string, includeDeleted bool) (*domain.PaperDetail, error) {
	record, err := r.queries.GetPaperByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !includeDeleted && record.DeletedAt.Valid {
		return nil, pgx.ErrNoRows
	}

	items, err := r.buildPaperDetails(ctx, []sqlc.Paper{record}, includeDeleted)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}

	return &items[0], nil
}

func (r *Repository) GetExportJob(ctx context.Context, id string) (*domain.ExportJob, error) {
	record, err := r.queries.GetExportJobByID(ctx, id)
	if err != nil {
		return nil, err
	}

	item := r.exportJobFromRecord(record)
	return &item, nil
}

func (r *Repository) GetImportJob(ctx context.Context, id string) (*domain.ImportJob, error) {
	record, err := r.queries.GetImportJobByID(ctx, id)
	if err != nil {
		return nil, err
	}

	item := r.importJobFromRecord(record)
	return &item, nil
}

func (r *Repository) ListProblems(ctx context.Context, opts ProblemListOptions) (domain.Paginated[domain.Problem], error) {
	page, pageSize, offset := normalizePage(opts.Page, opts.PageSize)
	predicates := buildProblemPredicates(problemPredicateOptions{
		IncludeDeleted: opts.IncludeDeleted,
		Keyword:        opts.Keyword,
		Subject:        opts.Subject,
		Grade:          opts.Grade,
		Difficulty:     opts.Difficulty,
		Type:           opts.Type,
		TagIDs:         opts.TagIDs,
		HasImage:       opts.HasImage,
		ScoreMin:       opts.ScoreMin,
		ScoreMax:       opts.ScoreMax,
	})

	countQuery := "SELECT count(*) FROM problems p " + predicates.clause
	var total int
	if err := r.db.QueryRow(ctx, countQuery, predicates.args...).Scan(&total); err != nil {
		return domain.Paginated[domain.Problem]{}, err
	}

	args := append([]any{}, predicates.args...)
	limitArg := fmt.Sprintf("$%d", len(args)+1)
	offsetArg := fmt.Sprintf("$%d", len(args)+2)
	args = append(args, pageSize, offset)

	query := `SELECT id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
  subjective_score, subject, grade, source, notes, search_tsv::text AS search_tsv, formula_tokens, formula_tsv::text AS formula_tsv,
  version, created_at, updated_at, deleted_at
FROM problems p ` + predicates.clause + `
ORDER BY ` + problemOrderClause(opts.SortBy, opts.SortOrder) + `
LIMIT ` + limitArg + ` OFFSET ` + offsetArg

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return domain.Paginated[domain.Problem]{}, err
	}
	defer rows.Close()

	records := make([]sqlc.Problem, 0)
	for rows.Next() {
		record, err := scanProblemRecord(rows)
		if err != nil {
			return domain.Paginated[domain.Problem]{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return domain.Paginated[domain.Problem]{}, err
	}

	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	tagMap, err := r.loadProblemTagIDsForProblems(ctx, ids)
	if err != nil {
		return domain.Paginated[domain.Problem]{}, err
	}
	imageMap, err := r.loadProblemImageIDsForProblems(ctx, ids)
	if err != nil {
		return domain.Paginated[domain.Problem]{}, err
	}

	items := make([]domain.Problem, 0, len(records))
	for _, record := range records {
		items = append(items, r.problemFromRecord(record, tagMap[record.ID], imageMap[record.ID]))
	}

	return domain.Paginated[domain.Problem]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *Repository) SearchProblems(ctx context.Context, opts SearchOptions) ([]domain.SearchResult, error) {
	predicates := buildProblemPredicates(problemPredicateOptions{
		Keyword:    opts.Keyword,
		Formula:    opts.Formula,
		Conditions: opts.Conditions,
	})

	args := append([]any{}, predicates.args...)
	snippetExpr := "p.latex"
	orderClause := "p.updated_at DESC"
	if predicates.keywordArg != "" {
		snippetExpr = fmt.Sprintf(
			`ts_headline('simple', p.latex, plainto_tsquery('simple', %s), 'StartSel=<mark>, StopSel=</mark>, MaxFragments=2, MinWords=3, MaxWords=20')`,
			predicates.keywordArg,
		)
		orderClause = fmt.Sprintf("ts_rank_cd(p.search_tsv, plainto_tsquery('simple', %s)) DESC, p.updated_at DESC", predicates.keywordArg)
	}

	query := `SELECT id, code, latex, answer_latex, solution_latex, problem_type, difficulty,
  subjective_score, subject, grade, source, notes, search_tsv::text AS search_tsv, formula_tokens, formula_tsv::text AS formula_tsv,
  version, created_at, updated_at, deleted_at, ` + snippetExpr + `
FROM problems p ` + predicates.clause + `
ORDER BY ` + orderClause

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := make([]searchRow, 0)
	for rows.Next() {
		record, snippet, err := scanSearchProblemRecord(rows)
		if err != nil {
			return nil, err
		}
		found = append(found, searchRow{Problem: record, Snippet: snippet})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	records := make([]sqlc.Problem, 0, len(found))
	snippets := make(map[string]string, len(found))
	for _, item := range found {
		records = append(records, item.Problem)
		snippets[item.Problem.ID] = item.Snippet
	}
	details, err := r.buildProblemDetails(ctx, records, false)
	if err != nil {
		return nil, err
	}

	results := make([]domain.SearchResult, 0, len(details))
	for _, detail := range details {
		results = append(results, domain.SearchResult{
			ProblemDetail: detail,
			Snippet:       snippets[detail.ID],
		})
	}
	return results, nil
}

func (r *Repository) ListImages(ctx context.Context, opts ImageListOptions) (domain.Paginated[domain.ImageAsset], error) {
	page, pageSize, offset := normalizePage(opts.Page, opts.PageSize)
	args := &sqlArgs{}
	conditions := make([]string, 0, 4)

	if !opts.IncludeDeleted {
		conditions = append(conditions, "i.deleted_at IS NULL")
	}
	if keyword := strings.TrimSpace(opts.Keyword); keyword != "" {
		value := args.Add("%" + strings.ToLower(keyword) + "%")
		conditions = append(conditions, fmt.Sprintf("(LOWER(i.filename) LIKE %s OR LOWER(COALESCE(i.description, '')) LIKE %s)", value, value))
	}
	if opts.MIME != "" {
		conditions = append(conditions, "i.mime = "+args.Add(opts.MIME))
	}
	if len(opts.TagIDs) > 0 {
		tagArg := args.Add(opts.TagIDs)
		countArg := args.Add(len(opts.TagIDs))
		conditions = append(conditions, fmt.Sprintf(`
i.id IN (
  SELECT it.image_id
  FROM image_tags it
  WHERE it.tag_id = ANY(%s::text[])
  GROUP BY it.image_id
  HAVING COUNT(DISTINCT it.tag_id) >= %s
)`, tagArg, countArg))
	}

	where := buildWhereClause(conditions)

	countQuery := "SELECT count(*) FROM images i " + where
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args.values...).Scan(&total); err != nil {
		return domain.Paginated[domain.ImageAsset]{}, err
	}

	queryArgs := append([]any{}, args.values...)
	limitArg := fmt.Sprintf("$%d", len(queryArgs)+1)
	offsetArg := fmt.Sprintf("$%d", len(queryArgs)+2)
	queryArgs = append(queryArgs, pageSize, offset)

	query := `SELECT id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path,
  description, parent_image_id, created_at, updated_at, deleted_at
FROM images i ` + where + `
ORDER BY i.created_at DESC
LIMIT ` + limitArg + ` OFFSET ` + offsetArg

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return domain.Paginated[domain.ImageAsset]{}, err
	}
	defer rows.Close()

	records := make([]sqlc.Image, 0)
	for rows.Next() {
		record, err := scanImageRecord(rows)
		if err != nil {
			return domain.Paginated[domain.ImageAsset]{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return domain.Paginated[domain.ImageAsset]{}, err
	}

	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
	}
	tagMap, err := r.loadImageTagIDsForImages(ctx, ids)
	if err != nil {
		return domain.Paginated[domain.ImageAsset]{}, err
	}
	linkedMap, err := r.loadImageLinkedProblemIDsForImages(ctx, ids)
	if err != nil {
		return domain.Paginated[domain.ImageAsset]{}, err
	}

	items := make([]domain.ImageAsset, 0, len(records))
	for _, record := range records {
		items = append(items, r.imageFromRecord(record, tagMap[record.ID], linkedMap[record.ID]))
	}

	return domain.Paginated[domain.ImageAsset]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *Repository) ListPapers(ctx context.Context, opts PaperListOptions) (domain.Paginated[domain.PaperDetail], error) {
	page, pageSize, offset := normalizePage(opts.Page, opts.PageSize)
	args := &sqlArgs{}
	conditions := make([]string, 0, 2)

	if !opts.IncludeDeleted {
		conditions = append(conditions, "p.deleted_at IS NULL")
	}
	if keyword := strings.TrimSpace(opts.Keyword); keyword != "" {
		value := args.Add("%" + strings.ToLower(keyword) + "%")
		conditions = append(conditions, fmt.Sprintf("(LOWER(p.title) LIKE %s OR LOWER(COALESCE(p.subtitle, '')) LIKE %s)", value, value))
	}

	where := buildWhereClause(conditions)
	countQuery := "SELECT count(*) FROM papers p " + where
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args.values...).Scan(&total); err != nil {
		return domain.Paginated[domain.PaperDetail]{}, err
	}

	queryArgs := append([]any{}, args.values...)
	limitArg := fmt.Sprintf("$%d", len(queryArgs)+1)
	offsetArg := fmt.Sprintf("$%d", len(queryArgs)+2)
	queryArgs = append(queryArgs, pageSize, offset)

	query := `SELECT id, title, subtitle, school_name, exam_name, subject, duration_min,
  total_score, description, status, instructions, footer_text, header_json, layout_json,
  created_at, updated_at, deleted_at
FROM papers p ` + where + `
ORDER BY p.updated_at DESC
LIMIT ` + limitArg + ` OFFSET ` + offsetArg

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return domain.Paginated[domain.PaperDetail]{}, err
	}
	defer rows.Close()

	records := make([]sqlc.Paper, 0)
	for rows.Next() {
		record, err := scanPaperRecord(rows)
		if err != nil {
			return domain.Paginated[domain.PaperDetail]{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return domain.Paginated[domain.PaperDetail]{}, err
	}

	items, err := r.buildPaperDetails(ctx, records, opts.IncludeDeleted)
	if err != nil {
		return domain.Paginated[domain.PaperDetail]{}, err
	}

	return domain.Paginated[domain.PaperDetail]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *Repository) ListExportJobs(ctx context.Context, opts ExportListOptions) (domain.Paginated[domain.ExportJob], error) {
	page, pageSize, offset := normalizePage(opts.Page, opts.PageSize)
	args := &sqlArgs{}
	where := ""
	if opts.Status != "" {
		where = "WHERE status = " + args.Add(opts.Status)
	}

	countQuery := "SELECT count(*) FROM export_jobs " + where
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args.values...).Scan(&total); err != nil {
		return domain.Paginated[domain.ExportJob]{}, err
	}

	queryArgs := append([]any{}, args.values...)
	limitArg := fmt.Sprintf("$%d", len(queryArgs)+1)
	offsetArg := fmt.Sprintf("$%d", len(queryArgs)+2)
	queryArgs = append(queryArgs, pageSize, offset)

	query := `SELECT id, paper_id, paper_title, format, variant, status, progress,
  download_path, error_message, created_at, started_at, completed_at, cancel_requested_at
FROM export_jobs ` + where + `
ORDER BY created_at DESC
LIMIT ` + limitArg + ` OFFSET ` + offsetArg

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return domain.Paginated[domain.ExportJob]{}, err
	}
	defer rows.Close()

	items := make([]domain.ExportJob, 0)
	for rows.Next() {
		record, err := scanExportRecord(rows)
		if err != nil {
			return domain.Paginated[domain.ExportJob]{}, err
		}
		items = append(items, r.exportJobFromRecord(record))
	}
	if err := rows.Err(); err != nil {
		return domain.Paginated[domain.ExportJob]{}, err
	}

	return domain.Paginated[domain.ExportJob]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *Repository) ListImportJobs(ctx context.Context, opts ImportListOptions) (domain.Paginated[domain.ImportJob], error) {
	page, pageSize, offset := normalizePage(opts.Page, opts.PageSize)
	args := &sqlArgs{}
	where := ""
	if opts.Status != "" {
		where = "WHERE status = " + args.Add(opts.Status)
	}

	countQuery := "SELECT count(*) FROM import_jobs " + where
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args.values...).Scan(&total); err != nil {
		return domain.Paginated[domain.ImportJob]{}, err
	}

	queryArgs := append([]any{}, args.values...)
	limitArg := fmt.Sprintf("$%d", len(queryArgs)+1)
	offsetArg := fmt.Sprintf("$%d", len(queryArgs)+2)
	queryArgs = append(queryArgs, pageSize, offset)

	query := `SELECT id, filename, input_type, status, progress, result, error_message,
  created_at, started_at, completed_at
FROM import_jobs ` + where + `
ORDER BY created_at DESC
LIMIT ` + limitArg + ` OFFSET ` + offsetArg

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return domain.Paginated[domain.ImportJob]{}, err
	}
	defer rows.Close()

	items := make([]domain.ImportJob, 0)
	for rows.Next() {
		record, err := scanImportRecord(rows)
		if err != nil {
			return domain.Paginated[domain.ImportJob]{}, err
		}
		items = append(items, r.importJobFromRecord(record))
	}
	if err := rows.Err(); err != nil {
		return domain.Paginated[domain.ImportJob]{}, err
	}

	return domain.Paginated[domain.ImportJob]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *Repository) ListTags(ctx context.Context) ([]domain.Tag, error) {
	rows, err := r.db.Query(ctx, `SELECT
  t.id,
  t.name,
  t.category,
  t.color,
  t.description,
  t.created_at,
  t.updated_at,
  COALESCE(COUNT(pt.problem_id) FILTER (WHERE p.deleted_at IS NULL), 0) AS problem_count
FROM tags t
LEFT JOIN problem_tags pt ON pt.tag_id = t.id
LEFT JOIN problems p ON p.id = pt.problem_id
GROUP BY t.id, t.name, t.category, t.color, t.description, t.created_at, t.updated_at
ORDER BY t.name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Tag, 0)
	for rows.Next() {
		record, count, err := scanTagRecord(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, r.tagFromRecord(record, count))
	}
	return items, rows.Err()
}

func (r *Repository) ListProblemSubjects(ctx context.Context) ([]string, error) {
	return r.listProblemMetadata(ctx, "subject")
}

func (r *Repository) ListProblemGrades(ctx context.Context) ([]string, error) {
	values, err := r.listProblemMetadata(ctx, "grade")
	if err != nil {
		return nil, err
	}
	return domain.BuildGradeOptions(values), nil
}

func (r *Repository) GetSettings(ctx context.Context) (domain.SettingsPayload, error) {
	return r.loadSettings(ctx)
}

func (r *Repository) MetaStats(ctx context.Context) (domain.MetaStats, error) {
	stats := domain.MetaStats{}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM problems WHERE deleted_at IS NULL`).Scan(&stats.ProblemCount); err != nil {
		return domain.MetaStats{}, err
	}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM problems WHERE deleted_at IS NULL AND created_at >= now() - interval '7 days'`).Scan(&stats.RecentProblemGain); err != nil {
		return domain.MetaStats{}, err
	}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM images WHERE deleted_at IS NULL`).Scan(&stats.ImageCount); err != nil {
		return domain.MetaStats{}, err
	}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM tags`).Scan(&stats.TagCount); err != nil {
		return domain.MetaStats{}, err
	}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM export_jobs WHERE created_at >= now() - interval '30 days'`).Scan(&stats.ExportCount); err != nil {
		return domain.MetaStats{}, err
	}
	return stats, nil
}

func (r *Repository) NotifyExportJobChanged(ctx context.Context, jobID string) error {
	job, err := r.GetExportJob(ctx, jobID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(exportNotificationPayload{
		JobID:    job.ID,
		Status:   string(job.Status),
		Progress: job.Progress,
	})
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `SELECT pg_notify('export_channel', $1)`, string(payload))
	return err
}

func (r *Repository) IsExportCancellationRequested(ctx context.Context, jobID string) (bool, error) {
	var requested bool
	err := r.db.QueryRow(ctx, `SELECT cancel_requested_at IS NOT NULL FROM export_jobs WHERE id = $1`, jobID).Scan(&requested)
	return requested, err
}

func (r *Repository) buildProblemDetails(ctx context.Context, records []sqlc.Problem, includeDeleted bool) ([]domain.ProblemDetail, error) {
	if len(records) == 0 {
		return []domain.ProblemDetail{}, nil
	}

	problemIDs := make([]string, 0, len(records))
	for _, record := range records {
		problemIDs = append(problemIDs, record.ID)
	}

	tagIDsByProblemID, err := r.loadProblemTagIDsForProblems(ctx, problemIDs)
	if err != nil {
		return nil, err
	}
	imageIDsByProblemID, err := r.loadProblemImageIDsForProblems(ctx, problemIDs)
	if err != nil {
		return nil, err
	}

	tagIDs := uniqueIDsFromMap(tagIDsByProblemID)
	imageIDs := uniqueIDsFromMap(imageIDsByProblemID)

	tagMap, err := r.loadTagMapByIDs(ctx, tagIDs)
	if err != nil {
		return nil, err
	}
	imageMap, err := r.loadImageAssetsByIDs(ctx, imageIDs, includeDeleted)
	if err != nil {
		return nil, err
	}

	items := make([]domain.ProblemDetail, 0, len(records))
	for _, record := range records {
		tagIDs := tagIDsByProblemID[record.ID]
		imageIDs := imageIDsByProblemID[record.ID]
		detail := domain.ProblemDetail{
			Problem: r.problemFromRecord(record, tagIDs, imageIDs),
			Tags:    make([]domain.Tag, 0, len(tagIDs)),
			Images:  make([]domain.ImageAsset, 0, len(imageIDs)),
		}
		for _, tagID := range tagIDs {
			if tag, ok := tagMap[tagID]; ok {
				detail.Tags = append(detail.Tags, tag)
			}
		}
		for _, imageID := range imageIDs {
			if imageAsset, ok := imageMap[imageID]; ok {
				detail.Images = append(detail.Images, imageAsset)
			}
		}
		items = append(items, detail)
	}

	return items, nil
}

func (r *Repository) buildPaperDetails(ctx context.Context, records []sqlc.Paper, includeDeleted bool) ([]domain.PaperDetail, error) {
	if len(records) == 0 {
		return []domain.PaperDetail{}, nil
	}

	paperIDs := make([]string, 0, len(records))
	for _, record := range records {
		paperIDs = append(paperIDs, record.ID)
	}

	paperItems, err := r.loadPaperItemsByPaperIDs(ctx, paperIDs)
	if err != nil {
		return nil, err
	}

	problemIDs := make([]string, 0)
	for _, items := range paperItems {
		for _, item := range items {
			problemIDs = append(problemIDs, item.ProblemID)
		}
	}

	problemDetails, err := r.loadProblemDetailsByIDs(ctx, problemIDs, includeDeleted)
	if err != nil {
		return nil, err
	}
	problemByID := make(map[string]domain.ProblemDetail, len(problemDetails))
	for _, detail := range problemDetails {
		problemByID[detail.ID] = detail
	}

	items := make([]domain.PaperDetail, 0, len(records))
	for _, record := range records {
		layout := domain.PaperLayout{
			Columns:           1,
			FontSize:          12,
			LineHeight:        1.3,
			PaperSize:         "A4",
			ShowAnswerVersion: true,
		}
		if len(record.LayoutJson) > 0 {
			if err := json.Unmarshal(record.LayoutJson, &layout); err != nil {
				return nil, fmt.Errorf("decode paper layout: %w", err)
			}
		}

		header := map[string]any{}
		if len(record.HeaderJson) > 0 {
			if err := json.Unmarshal(record.HeaderJson, &header); err != nil {
				return nil, fmt.Errorf("decode paper header: %w", err)
			}
		}

		detail := domain.PaperDetail{
			Paper: domain.Paper{
				ID:         record.ID,
				Title:      record.Title,
				Subtitle:   pgTextPtr(record.Subtitle),
				SchoolName: pgTextPtr(record.SchoolName),
				ExamName:   pgTextPtr(record.ExamName),
				Subject:    pgTextPtr(record.Subject),
				Duration:   pgInt4Ptr(record.DurationMin),
				TotalScore: pgNumericPtr(record.TotalScore),
				Items:      make([]domain.PaperItem, 0),
				Layout:     layout,
				CreatedAt:  pgTimestamptzValue(record.CreatedAt),
				UpdatedAt:  pgTimestamptzValue(record.UpdatedAt),
			},
			Description:  pgTextPtr(record.Description),
			Status:       domain.PaperStatus(record.Status),
			Instructions: pgTextPtr(record.Instructions),
			FooterText:   pgTextPtr(record.FooterText),
			Header:       header,
			ItemDetails:  make([]domain.PaperItemDetail, 0),
		}

		for _, itemRecord := range paperItems[record.ID] {
			item := domain.PaperItem{
				ID:            itemRecord.ID,
				ProblemID:     itemRecord.ProblemID,
				Score:         pgNumericValue(itemRecord.Score),
				OrderIndex:    int(itemRecord.OrderIndex),
				ImagePosition: pgTextValue(itemRecord.ImagePosition, "below"),
				BlankLines:    int(itemRecord.BlankLines),
			}
			detail.Items = append(detail.Items, item)

			itemDetail := domain.PaperItemDetail{PaperItem: item}
			if problem, ok := problemByID[item.ProblemID]; ok {
				copyProblem := problem
				itemDetail.Problem = &copyProblem
			}
			detail.ItemDetails = append(detail.ItemDetails, itemDetail)
		}

		items = append(items, detail)
	}

	return items, nil
}

func (r *Repository) loadProblemDetailsByIDs(ctx context.Context, ids []string, includeDeleted bool) ([]domain.ProblemDetail, error) {
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return []domain.ProblemDetail{}, nil
	}

	records, err := r.queries.GetProblemsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	filtered := make([]sqlc.Problem, 0, len(records))
	for _, record := range records {
		if !includeDeleted && record.DeletedAt.Valid {
			continue
		}
		filtered = append(filtered, problemRecordFromGetProblemsByIDsRow(record))
	}

	return r.buildProblemDetails(ctx, filtered, includeDeleted)
}

func (r *Repository) loadImageAssetsByIDs(ctx context.Context, ids []string, includeDeleted bool) (map[string]domain.ImageAsset, error) {
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return map[string]domain.ImageAsset{}, nil
	}

	query := `SELECT id, filename, mime, size_bytes, width, height, storage_path, thumbnail_path,
  description, parent_image_id, created_at, updated_at, deleted_at
FROM images
WHERE id = ANY($1::text[])`
	if !includeDeleted {
		query += ` AND deleted_at IS NULL`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]sqlc.Image, 0)
	for rows.Next() {
		record, err := scanImageRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	imageIDs := make([]string, 0, len(records))
	for _, record := range records {
		imageIDs = append(imageIDs, record.ID)
	}
	tagMap, err := r.loadImageTagIDsForImages(ctx, imageIDs)
	if err != nil {
		return nil, err
	}
	linkedMap, err := r.loadImageLinkedProblemIDsForImages(ctx, imageIDs)
	if err != nil {
		return nil, err
	}

	items := make(map[string]domain.ImageAsset, len(records))
	for _, record := range records {
		items[record.ID] = r.imageFromRecord(record, tagMap[record.ID], linkedMap[record.ID])
	}
	return items, nil
}

func (r *Repository) loadPaperItemsByPaperIDs(ctx context.Context, ids []string) (map[string][]sqlc.PaperItem, error) {
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return map[string][]sqlc.PaperItem{}, nil
	}

	rows, err := r.db.Query(ctx, `SELECT id, paper_id, problem_id, order_index, score, image_position, blank_lines
FROM paper_items
WHERE paper_id = ANY($1::text[])
ORDER BY paper_id, order_index`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := map[string][]sqlc.PaperItem{}
	for rows.Next() {
		var item sqlc.PaperItem
		if err := rows.Scan(&item.ID, &item.PaperID, &item.ProblemID, &item.OrderIndex, &item.Score, &item.ImagePosition, &item.BlankLines); err != nil {
			return nil, err
		}
		items[item.PaperID] = append(items[item.PaperID], item)
	}
	return items, rows.Err()
}

func (r *Repository) loadProblemTagIDsForProblems(ctx context.Context, ids []string) (map[string][]string, error) {
	return r.loadRelations(ctx, `SELECT problem_id, tag_id FROM problem_tags WHERE problem_id = ANY($1::text[])`, ids)
}

func (r *Repository) loadProblemImageIDsForProblems(ctx context.Context, ids []string) (map[string][]string, error) {
	return r.loadRelations(ctx, `SELECT problem_id, image_id FROM problem_images WHERE problem_id = ANY($1::text[]) ORDER BY problem_id, order_index`, ids)
}

func (r *Repository) loadImageTagIDsForImages(ctx context.Context, ids []string) (map[string][]string, error) {
	return r.loadRelations(ctx, `SELECT image_id, tag_id FROM image_tags WHERE image_id = ANY($1::text[])`, ids)
}

func (r *Repository) loadImageLinkedProblemIDsForImages(ctx context.Context, ids []string) (map[string][]string, error) {
	return r.loadRelations(ctx, `SELECT image_id, problem_id FROM problem_images WHERE image_id = ANY($1::text[]) ORDER BY image_id, order_index`, ids)
}

func (r *Repository) loadRelations(ctx context.Context, query string, ids []string) (map[string][]string, error) {
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return map[string][]string{}, nil
	}

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := map[string][]string{}
	for rows.Next() {
		var ownerID string
		var relatedID string
		if err := rows.Scan(&ownerID, &relatedID); err != nil {
			return nil, err
		}
		items[ownerID] = append(items[ownerID], relatedID)
	}
	return items, rows.Err()
}

func (r *Repository) loadTagMapByIDs(ctx context.Context, ids []string) (map[string]domain.Tag, error) {
	ids = uniqueIDs(ids)
	if len(ids) == 0 {
		return map[string]domain.Tag{}, nil
	}

	rows, err := r.db.Query(ctx, `SELECT
  t.id,
  t.name,
  t.category,
  t.color,
  t.description,
  t.created_at,
  t.updated_at,
  COALESCE(COUNT(pt.problem_id) FILTER (WHERE p.deleted_at IS NULL), 0) AS problem_count
FROM tags t
LEFT JOIN problem_tags pt ON pt.tag_id = t.id
LEFT JOIN problems p ON p.id = pt.problem_id
WHERE t.id = ANY($1::text[])
GROUP BY t.id, t.name, t.category, t.color, t.description, t.created_at, t.updated_at`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(map[string]domain.Tag, len(ids))
	for rows.Next() {
		record, count, err := scanTagRecord(rows)
		if err != nil {
			return nil, err
		}
		items[record.ID] = r.tagFromRecord(record, count)
	}
	return items, rows.Err()
}

var metadataColumns = map[string]string{
	"subject": "subject",
	"grade":   "grade",
}

func (r *Repository) listProblemMetadata(ctx context.Context, column string) ([]string, error) {
	col, ok := metadataColumns[column]
	if !ok {
		return nil, fmt.Errorf("unsupported metadata column: %s", column)
	}

	rows, err := r.db.Query(ctx, `SELECT DISTINCT `+col+` FROM problems WHERE deleted_at IS NULL AND `+col+` IS NOT NULL AND `+col+` <> '' ORDER BY `+col+` ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make([]string, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

func (r *Repository) problemFromRecord(record sqlc.Problem, tagIDs, imageIDs []string) domain.Problem {
	return domain.Problem{
		ID:              record.ID,
		Code:            record.Code,
		Latex:           record.Latex,
		AnswerLatex:     pgTextPtr(record.AnswerLatex),
		SolutionLatex:   pgTextPtr(record.SolutionLatex),
		Type:            domain.ProblemType(record.ProblemType),
		Difficulty:      domain.Difficulty(record.Difficulty),
		SubjectiveScore: pgNumericPtr(record.SubjectiveScore),
		Subject:         pgTextPtr(record.Subject),
		Grade:           pgTextPtr(record.Grade),
		Source:          pgTextPtr(record.Source),
		TagIDs:          copyStrings(tagIDs),
		ImageIDs:        copyStrings(imageIDs),
		Notes:           pgTextPtr(record.Notes),
		CreatedAt:       pgTimestamptzValue(record.CreatedAt),
		UpdatedAt:       pgTimestamptzValue(record.UpdatedAt),
		Version:         int(record.Version),
		IsDeleted:       record.DeletedAt.Valid,
	}
}

func (r *Repository) imageFromRecord(record sqlc.Image, tagIDs, linkedProblemIDs []string) domain.ImageAsset {
	return domain.ImageAsset{
		ID:               record.ID,
		Filename:         record.Filename,
		MIME:             record.Mime,
		Size:             record.SizeBytes,
		Width:            int(record.Width),
		Height:           int(record.Height),
		URL:              fmt.Sprintf("%s/api/v1/images/%s/file", r.publicBaseURL, record.ID),
		ThumbnailURL:     fmt.Sprintf("%s/api/v1/images/%s/thumbnail", r.publicBaseURL, record.ID),
		TagIDs:           copyStrings(tagIDs),
		LinkedProblemIDs: copyStrings(linkedProblemIDs),
		Description:      pgTextPtr(record.Description),
		CreatedAt:        pgTimestamptzValue(record.CreatedAt),
		UpdatedAt:        pgTimestamptzValue(record.UpdatedAt),
		IsDeleted:        record.DeletedAt.Valid,
	}
}

func (r *Repository) exportJobFromRecord(record sqlc.ExportJob) domain.ExportJob {
	item := domain.ExportJob{
		ID:           record.ID,
		PaperID:      record.PaperID,
		PaperTitle:   record.PaperTitle,
		Format:       domain.ExportFormat(record.Format),
		Variant:      domain.ExportVariant(record.Variant),
		Status:       domain.ExportStatus(record.Status),
		Progress:     int(record.Progress),
		ErrorMessage: pgTextPtr(record.ErrorMessage),
		CreatedAt:    pgTimestamptzValue(record.CreatedAt),
		StartedAt:    pgTimestamptzPtr(record.StartedAt),
		CompletedAt:  pgTimestamptzPtr(record.CompletedAt),
	}
	if record.DownloadPath.Valid {
		url := fmt.Sprintf("%s/api/v1/exports/%s/download", r.publicBaseURL, record.ID)
		item.DownloadURL = &url
	}
	return item
}

func (r *Repository) tagFromRecord(record sqlc.Tag, problemCount int) domain.Tag {
	return domain.Tag{
		ID:           record.ID,
		Name:         record.Name,
		Category:     domain.TagCategory(record.Category),
		Color:        record.Color,
		Description:  pgTextPtr(record.Description),
		ProblemCount: problemCount,
	}
}

func scanProblemRecord(scanner rowScanner) (sqlc.Problem, error) {
	var record sqlc.Problem
	err := scanner.Scan(
		&record.ID,
		&record.Code,
		&record.Latex,
		&record.AnswerLatex,
		&record.SolutionLatex,
		&record.ProblemType,
		&record.Difficulty,
		&record.SubjectiveScore,
		&record.Subject,
		&record.Grade,
		&record.Source,
		&record.Notes,
		&record.SearchTsv,
		&record.FormulaTokens,
		&record.FormulaTsv,
		&record.Version,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.DeletedAt,
	)
	return record, err
}

func scanImageRecord(scanner rowScanner) (sqlc.Image, error) {
	var record sqlc.Image
	err := scanner.Scan(
		&record.ID,
		&record.Filename,
		&record.Mime,
		&record.SizeBytes,
		&record.Width,
		&record.Height,
		&record.StoragePath,
		&record.ThumbnailPath,
		&record.Description,
		&record.ParentImageID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.DeletedAt,
	)
	return record, err
}

func scanPaperRecord(scanner rowScanner) (sqlc.Paper, error) {
	var record sqlc.Paper
	err := scanner.Scan(
		&record.ID,
		&record.Title,
		&record.Subtitle,
		&record.SchoolName,
		&record.ExamName,
		&record.Subject,
		&record.DurationMin,
		&record.TotalScore,
		&record.Description,
		&record.Status,
		&record.Instructions,
		&record.FooterText,
		&record.HeaderJson,
		&record.LayoutJson,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.DeletedAt,
	)
	return record, err
}

func scanExportRecord(scanner rowScanner) (sqlc.ExportJob, error) {
	var record sqlc.ExportJob
	err := scanner.Scan(
		&record.ID,
		&record.PaperID,
		&record.PaperTitle,
		&record.Format,
		&record.Variant,
		&record.Status,
		&record.Progress,
		&record.DownloadPath,
		&record.ErrorMessage,
		&record.CreatedAt,
		&record.StartedAt,
		&record.CompletedAt,
		&record.CancelRequestedAt,
	)
	return record, err
}

func scanImportRecord(scanner rowScanner) (sqlc.ImportJob, error) {
	var record sqlc.ImportJob
	err := scanner.Scan(
		&record.ID,
		&record.Filename,
		&record.InputType,
		&record.Status,
		&record.Progress,
		&record.Result,
		&record.ErrorMessage,
		&record.CreatedAt,
		&record.StartedAt,
		&record.CompletedAt,
	)
	return record, err
}

func scanSearchProblemRecord(scanner rowScanner) (sqlc.Problem, string, error) {
	record, err := scanProblemRecordWithSnippet(scanner, nil)
	return record.Problem, record.Snippet, err
}

type scannedProblemWithSnippet struct {
	Problem sqlc.Problem
	Snippet string
}

func scanProblemRecordWithSnippet(scanner rowScanner, snippet *string) (scannedProblemWithSnippet, error) {
	var (
		record scannedProblemWithSnippet
		target any = new(string)
	)
	if snippet != nil {
		target = snippet
	}
	err := scanner.Scan(
		&record.Problem.ID,
		&record.Problem.Code,
		&record.Problem.Latex,
		&record.Problem.AnswerLatex,
		&record.Problem.SolutionLatex,
		&record.Problem.ProblemType,
		&record.Problem.Difficulty,
		&record.Problem.SubjectiveScore,
		&record.Problem.Subject,
		&record.Problem.Grade,
		&record.Problem.Source,
		&record.Problem.Notes,
		&record.Problem.SearchTsv,
		&record.Problem.FormulaTokens,
		&record.Problem.FormulaTsv,
		&record.Problem.Version,
		&record.Problem.CreatedAt,
		&record.Problem.UpdatedAt,
		&record.Problem.DeletedAt,
		target,
	)
	if err != nil {
		return scannedProblemWithSnippet{}, err
	}
	switch value := target.(type) {
	case *string:
		record.Snippet = *value
	}
	return record, nil
}

func scanTagRecord(scanner rowScanner) (sqlc.Tag, int, error) {
	var (
		record sqlc.Tag
		count  int64
	)
	err := scanner.Scan(
		&record.ID,
		&record.Name,
		&record.Category,
		&record.Color,
		&record.Description,
		&record.CreatedAt,
		&record.UpdatedAt,
		&count,
	)
	return record, int(count), err
}

func buildProblemPredicates(opts problemPredicateOptions) problemPredicateState {
	args := &sqlArgs{}
	conditions := make([]string, 0, 12)
	state := problemPredicateState{}

	if !opts.IncludeDeleted {
		conditions = append(conditions, "p.deleted_at IS NULL")
	}
	if keyword := strings.TrimSpace(opts.Keyword); keyword != "" {
		state.keywordArg = args.Add(keyword)
		// Use ILIKE for Chinese keywords (plainto_tsquery doesn't work well with Chinese)
		// For keywords containing only ASCII letters/numbers, use plainto_tsquery
		isAscii := func(s string) bool {
			for _, c := range s {
				if c >= 128 {
					return false
				}
			}
			return true
		}
		if isAscii(keyword) {
			conditions = append(conditions, "p.search_tsv @@ plainto_tsquery('simple', "+state.keywordArg+")")
		} else {
			conditions = append(conditions, "p.search_tsv::text ILIKE '%' || "+state.keywordArg+" || '%'")
		}
	}
	if formula := normalizeFormulaQuery(opts.Formula); formula != "" {
		conditions = append(conditions, "p.formula_tsv @@ to_tsquery('simple', "+args.Add(formula)+")")
	}
	if opts.Subject != "" {
		conditions = append(conditions, "p.subject = "+args.Add(opts.Subject))
	}
	if opts.Grade != "" {
		conditions = append(conditions, "p.grade = "+args.Add(opts.Grade))
	}
	if len(opts.Difficulty) > 0 {
		conditions = append(conditions, "p.difficulty = ANY("+args.Add(opts.Difficulty)+"::text[])")
	}
	if opts.Type != "" && opts.Type != "all" {
		conditions = append(conditions, "p.problem_type = "+args.Add(opts.Type))
	}
	if len(opts.TagIDs) > 0 {
		tagArg := args.Add(opts.TagIDs)
		countArg := args.Add(len(opts.TagIDs))
		conditions = append(conditions, fmt.Sprintf(`
p.id IN (
  SELECT pt.problem_id
  FROM problem_tags pt
  WHERE pt.tag_id = ANY(%s::text[])
  GROUP BY pt.problem_id
  HAVING COUNT(DISTINCT pt.tag_id) >= %s
)`, tagArg, countArg))
	}
	if opts.HasImage != nil {
		if *opts.HasImage {
			conditions = append(conditions, "EXISTS (SELECT 1 FROM problem_images pi WHERE pi.problem_id = p.id)")
		} else {
			conditions = append(conditions, "NOT EXISTS (SELECT 1 FROM problem_images pi WHERE pi.problem_id = p.id)")
		}
	}
	if opts.ScoreMin != nil {
		conditions = append(conditions, "p.subjective_score >= "+args.Add(*opts.ScoreMin))
	}
	if opts.ScoreMax != nil {
		conditions = append(conditions, "p.subjective_score <= "+args.Add(*opts.ScoreMax))
	}

	for _, condition := range opts.Conditions {
		switch condition.Field {
		case "subject":
			conditions = append(conditions, "p.subject = "+args.Add(condition.Value))
		case "grade":
			conditions = append(conditions, "p.grade = "+args.Add(condition.Value))
		case "difficulty":
			conditions = append(conditions, "p.difficulty = "+args.Add(condition.Value))
		case "type":
			conditions = append(conditions, "p.problem_type = "+args.Add(condition.Value))
		case "source":
			value := condition.Value
			if condition.Operator == "contains" {
				value = "%" + strings.ToLower(condition.Value) + "%"
				conditions = append(conditions, "LOWER(COALESCE(p.source, '')) LIKE "+args.Add(value))
			} else {
				conditions = append(conditions, "p.source = "+args.Add(value))
			}
		case "hasImage":
			if condition.Value == "yes" {
				conditions = append(conditions, "EXISTS (SELECT 1 FROM problem_images pi WHERE pi.problem_id = p.id)")
			}
			if condition.Value == "no" {
				conditions = append(conditions, "NOT EXISTS (SELECT 1 FROM problem_images pi WHERE pi.problem_id = p.id)")
			}
		case "tag":
			conditions = append(conditions, "EXISTS (SELECT 1 FROM problem_tags pt WHERE pt.problem_id = p.id AND pt.tag_id = "+args.Add(condition.Value)+")")
		case "subjectiveScore":
			switch condition.Operator {
			case "gt":
				conditions = append(conditions, "p.subjective_score > "+args.Add(condition.Value))
			case "lt":
				conditions = append(conditions, "p.subjective_score < "+args.Add(condition.Value))
			case "between":
				if condition.SecondValue != nil {
					conditions = append(conditions, "p.subjective_score BETWEEN "+args.Add(condition.Value)+" AND "+args.Add(*condition.SecondValue))
				}
			default:
				conditions = append(conditions, "p.subjective_score = "+args.Add(condition.Value))
			}
		case "date":
			switch condition.Operator {
			case "gt":
				conditions = append(conditions, "p.created_at::date > "+args.Add(condition.Value)+"::date")
			case "lt":
				conditions = append(conditions, "p.created_at::date < "+args.Add(condition.Value)+"::date")
			case "between":
				if condition.SecondValue != nil {
					conditions = append(conditions, "p.created_at::date BETWEEN "+args.Add(condition.Value)+"::date AND "+args.Add(*condition.SecondValue)+"::date")
				}
			default:
				conditions = append(conditions, "p.created_at::date = "+args.Add(condition.Value)+"::date")
			}
		}
	}

	state.clause = buildWhereClause(conditions)
	state.args = args.values
	return state
}

func buildWhereClause(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, "\n  AND ")
}

func problemOrderClause(sortBy, sortOrder string) string {
	column := "p.updated_at"
	switch strings.ToLower(sortBy) {
	case "code":
		column = "p.code"
	case "created_at":
		column = "p.created_at"
	case "updated_at":
		column = "p.updated_at"
	}

	direction := "DESC"
	if strings.ToLower(sortOrder) == "asc" {
		direction = "ASC"
	}
	return column + " " + direction
}

func normalizeFormulaQuery(value string) string {
	tokens := strings.Fields(search.TokenizeLatex(value))
	if len(tokens) == 0 {
		return ""
	}
	return strings.Join(tokens, " & ")
}

func normalizePage(page, pageSize int) (int, int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 24
	}
	return page, pageSize, (page - 1) * pageSize
}

func uniqueIDs(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(items))
	unique := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		unique = append(unique, item)
	}
	return unique
}

func uniqueIDsFromMap(items map[string][]string) []string {
	merged := make([]string, 0)
	for _, values := range items {
		merged = append(merged, values...)
	}
	return uniqueIDs(merged)
}
