package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/search"
	"mathlib/server/internal/store/sqlc"
)

func (r *Repository) CreateProblem(ctx context.Context, input domain.ProblemWriteInput) (*domain.ProblemDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	problemID, err := r.createProblemTx(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetProblemDetail(ctx, problemID, true)
}

func (r *Repository) createProblemTx(ctx context.Context, tx pgx.Tx, input domain.ProblemWriteInput) (string, error) {
	id := newID()
	code, err := r.NextProblemCode(ctx, tx)
	if err != nil {
		return "", err
	}
	queries := r.queries.WithTx(tx)
	now := time.Now().UTC()
	formulaTokens := search.TokenizeLatex(input.Latex)
	subjectiveScore, err := pgNumericFromPtr(input.SubjectiveScore)
	if err != nil {
		return "", err
	}
	if _, err := queries.InsertProblem(ctx, sqlc.InsertProblemParams{
		ID:              id,
		Code:            code,
		Latex:           input.Latex,
		AnswerLatex:     pgTextFromPtr(input.AnswerLatex),
		SolutionLatex:   pgTextFromPtr(input.SolutionLatex),
		ProblemType:     string(input.Type),
		Difficulty:      string(input.Difficulty),
		SubjectiveScore: subjectiveScore,
		Subject:         pgTextFromPtr(input.Subject),
		Grade:           pgTextFromPtr(input.Grade),
		Source:          pgTextFromPtr(input.Source),
		Notes:           pgTextFromPtr(input.Notes),
		FormulaTokens:   pgTextFromString(formulaTokens),
		Version:         1,
		CreatedAt:       pgTimestamptzFromTime(now),
		UpdatedAt:       pgTimestamptzFromTime(now),
	}); err != nil {
		return "", err
	}

	if err := r.replaceProblemRelations(ctx, queries, id, input.TagIDs, input.ImageIDs); err != nil {
		return "", err
	}
	if err := r.insertProblemVersion(ctx, queries, id, 1, input, code, now, false); err != nil {
		return "", err
	}

	return id, nil
}

func (r *Repository) UpdateProblem(ctx context.Context, id string, input domain.ProblemWriteInput) (*domain.ProblemDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	queries := r.queries.WithTx(tx)
	current, err := queries.GetProblemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	code := current.Code
	version := int(current.Version) + 1
	now := time.Now().UTC()
	formulaTokens := search.TokenizeLatex(input.Latex)
	subjectiveScore, err := pgNumericFromPtr(input.SubjectiveScore)
	if err != nil {
		return nil, err
	}
	if _, err := queries.UpdateProblem(ctx, sqlc.UpdateProblemParams{
		ID:              id,
		Latex:           input.Latex,
		AnswerLatex:     pgTextFromPtr(input.AnswerLatex),
		SolutionLatex:   pgTextFromPtr(input.SolutionLatex),
		ProblemType:     string(input.Type),
		Difficulty:      string(input.Difficulty),
		SubjectiveScore: subjectiveScore,
		Subject:         pgTextFromPtr(input.Subject),
		Grade:           pgTextFromPtr(input.Grade),
		Source:          pgTextFromPtr(input.Source),
		Notes:           pgTextFromPtr(input.Notes),
		FormulaTokens:   pgTextFromString(formulaTokens),
		Version:         int32(version),
		UpdatedAt:       pgTimestamptzFromTime(now),
	}); err != nil {
		return nil, err
	}

	if err := r.replaceProblemRelations(ctx, queries, id, input.TagIDs, input.ImageIDs); err != nil {
		return nil, err
	}
	if err := r.insertProblemVersion(ctx, queries, id, version, input, code, now, false); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetProblemDetail(ctx, id, true)
}

func (r *Repository) replaceProblemRelations(ctx context.Context, queries *sqlc.Queries, problemID string, tagIDs, imageIDs []string) error {
	if err := queries.DeleteProblemTags(ctx, problemID); err != nil {
		return err
	}
	if err := queries.DeleteProblemImages(ctx, problemID); err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		if err := queries.InsertProblemTag(ctx, sqlc.InsertProblemTagParams{ProblemID: problemID, TagID: tagID}); err != nil {
			return err
		}
	}

	for orderIndex, imageID := range imageIDs {
		if err := queries.InsertProblemImage(ctx, sqlc.InsertProblemImageParams{
			ProblemID:  problemID,
			ImageID:    imageID,
			OrderIndex: int32(orderIndex),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) insertProblemVersion(ctx context.Context, queries *sqlc.Queries, problemID string, version int, input domain.ProblemWriteInput, code string, timestamp time.Time, isDeleted bool) error {
	snapshot := map[string]any{
		"id":              problemID,
		"code":            code,
		"latex":           input.Latex,
		"answerLatex":     input.AnswerLatex,
		"solutionLatex":   input.SolutionLatex,
		"type":            input.Type,
		"difficulty":      input.Difficulty,
		"subjectiveScore": input.SubjectiveScore,
		"subject":         input.Subject,
		"grade":           input.Grade,
		"source":          input.Source,
		"tagIds":          input.TagIDs,
		"imageIds":        input.ImageIDs,
		"notes":           input.Notes,
		"updatedAt":       timestamp,
		"version":         version,
		"isDeleted":       isDeleted,
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return queries.InsertProblemVersion(ctx, sqlc.InsertProblemVersionParams{
		ID:        fmt.Sprintf("%s-v%d", problemID, version),
		ProblemID: problemID,
		Version:   int32(version),
		Snapshot:  raw,
		CreatedAt: pgTimestamptzFromTime(timestamp),
	})
}

func (r *Repository) SoftDeleteProblem(ctx context.Context, id string) error {
	return r.queries.SoftDeleteProblem(ctx, id)
}

func (r *Repository) RestoreProblem(ctx context.Context, id string) error {
	return r.queries.RestoreProblem(ctx, id)
}

func (r *Repository) HardDeleteProblem(ctx context.Context, id string) error {
	return r.queries.HardDeleteProblem(ctx, id)
}

func (r *Repository) ListProblemVersions(ctx context.Context, problemID string) ([]domain.ProblemVersion, error) {
	records, err := r.queries.GetProblemVersions(ctx, problemID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.ProblemVersion, 0, len(records))
	for _, record := range records {
		item := domain.ProblemVersion{
			ID:        record.ID,
			ProblemID: record.ProblemID,
			Version:   int(record.Version),
			CreatedAt: pgTimestamptzValue(record.CreatedAt),
		}
		var decoded any
		if err := json.Unmarshal(record.Snapshot, &decoded); err != nil {
			return nil, err
		}
		item.Snapshot = decoded
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) RollbackProblemVersion(ctx context.Context, problemID string, version int) (*domain.ProblemDetail, error) {
	var raw []byte
	if err := r.db.QueryRow(ctx, `SELECT snapshot FROM problem_versions WHERE problem_id = $1 AND version = $2`, problemID, version).Scan(&raw); err != nil {
		return nil, err
	}
	var snapshot struct {
		Latex           string             `json:"latex"`
		AnswerLatex     *string            `json:"answerLatex"`
		SolutionLatex   *string            `json:"solutionLatex"`
		Type            domain.ProblemType `json:"type"`
		Difficulty      domain.Difficulty  `json:"difficulty"`
		SubjectiveScore *float64           `json:"subjectiveScore"`
		Subject         *string            `json:"subject"`
		Grade           *string            `json:"grade"`
		Source          *string            `json:"source"`
		TagIDs          []string           `json:"tagIds"`
		ImageIDs        []string           `json:"imageIds"`
		Notes           *string            `json:"notes"`
	}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return r.UpdateProblem(ctx, problemID, domain.ProblemWriteInput{
		Latex:           snapshot.Latex,
		AnswerLatex:     snapshot.AnswerLatex,
		SolutionLatex:   snapshot.SolutionLatex,
		Type:            snapshot.Type,
		Difficulty:      snapshot.Difficulty,
		SubjectiveScore: snapshot.SubjectiveScore,
		Subject:         snapshot.Subject,
		Grade:           snapshot.Grade,
		Source:          snapshot.Source,
		TagIDs:          snapshot.TagIDs,
		ImageIDs:        snapshot.ImageIDs,
		Notes:           snapshot.Notes,
	})
}

func (r *Repository) CreateProblemsTx(ctx context.Context, tx pgx.Tx, inputs []domain.ProblemWriteInput) ([]string, error) {
	ids := make([]string, 0, len(inputs))
	queries := r.queries.WithTx(tx)

	for _, input := range inputs {
		id := newID()
		code, err := r.NextProblemCode(ctx, tx)
		if err != nil {
			return nil, err
		}
		now := time.Now().UTC()
		formulaTokens := search.TokenizeLatex(input.Latex)
		subjectiveScore, err := pgNumericFromPtr(input.SubjectiveScore)
		if err != nil {
			return nil, err
		}
		if _, err := queries.InsertProblem(ctx, sqlc.InsertProblemParams{
			ID:              id,
			Code:            code,
			Latex:           input.Latex,
			AnswerLatex:     pgTextFromPtr(input.AnswerLatex),
			SolutionLatex:   pgTextFromPtr(input.SolutionLatex),
			ProblemType:     string(input.Type),
			Difficulty:      string(input.Difficulty),
			SubjectiveScore: subjectiveScore,
			Subject:         pgTextFromPtr(input.Subject),
			Grade:           pgTextFromPtr(input.Grade),
			Source:          pgTextFromPtr(input.Source),
			Notes:           pgTextFromPtr(input.Notes),
			FormulaTokens:   pgTextFromString(formulaTokens),
			Version:         1,
			CreatedAt:       pgTimestamptzFromTime(now),
			UpdatedAt:       pgTimestamptzFromTime(now),
		}); err != nil {
			return nil, err
		}

		if err := r.replaceProblemRelations(ctx, queries, id, input.TagIDs, input.ImageIDs); err != nil {
			return nil, err
		}
		if err := r.insertProblemVersion(ctx, queries, id, 1, input, code, now, false); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *Repository) BatchTagProblems(ctx context.Context, problemIDs, tagIDs []string, replace bool) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, problemID := range problemIDs {
		if replace {
			if _, err := tx.Exec(ctx, `DELETE FROM problem_tags WHERE problem_id = $1`, problemID); err != nil {
				return err
			}
		}
		for _, tagID := range tagIDs {
			if _, err := tx.Exec(ctx, `INSERT INTO problem_tags (problem_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, problemID, tagID); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) BatchDeleteProblems(ctx context.Context, problemIDs []string) error {
	for _, problemID := range problemIDs {
		if err := r.SoftDeleteProblem(ctx, problemID); err != nil {
			return err
		}
	}
	return nil
}
