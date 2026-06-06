package store

import (
	"context"
	"encoding/json"
	"time"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store/sqlc"
)

func (r *Repository) CreateImportJobRecord(ctx context.Context, filename string, inputType string) (*domain.ImportJob, error) {
	id := newID()
	now := time.Now().UTC()
	record, err := r.queries.InsertImportJob(ctx, sqlc.InsertImportJobParams{
		ID:        id,
		Filename:  pgTextFromPtr(&filename),
		InputType: inputType,
		Status:    string(domain.ImportJobStatusPending),
		Progress:  0,
		CreatedAt: pgTimestamptzFromTime(now),
	})
	if err != nil {
		return nil, err
	}
	item := r.importJobFromRecord(record)
	return &item, nil
}

func (r *Repository) UpdateImportJobState(ctx context.Context, id string, status domain.ImportJobStatus, progress int, result []byte, errorMessage *string) error {
	current, err := r.queries.GetImportJobByID(ctx, id)
	if err != nil {
		return err
	}

	startedAt := current.StartedAt
	completedAt := current.CompletedAt
	resultData := current.Result
	errMessage := current.ErrorMessage

	if status == domain.ImportJobStatusProcessing {
		if !startedAt.Valid {
			startedAt = pgTimestamptzFromTime(time.Now().UTC())
		}
	}
	if status == domain.ImportJobStatusDone || status == domain.ImportJobStatusFailed {
		completedAt = pgTimestamptzFromTime(time.Now().UTC())
	}
	if result != nil {
		resultData = result
	}
	if errorMessage != nil {
		errMessage = pgTextFromPtr(errorMessage)
	}

	return r.queries.UpdateImportJobStatus(ctx, sqlc.UpdateImportJobStatusParams{
		ID:           id,
		Status:       string(status),
		Progress:     int32(progress),
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
		ErrorMessage: errMessage,
		Result:       resultData,
	})
}

func (r *Repository) DeleteImportJob(ctx context.Context, id string) error {
	return r.queries.DeleteImportJob(ctx, id)
}

func (r *Repository) NotifyImportJobChanged(ctx context.Context, jobID string) error {
	job, err := r.GetImportJob(ctx, jobID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(importNotificationPayload{
		JobID:    job.ID,
		Status:   string(job.Status),
		Progress: job.Progress,
	})
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `SELECT pg_notify('import_channel', $1)`, string(payload))
	return err
}

type importNotificationPayload struct {
	JobID    string `json:"jobId"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
}

func (r *Repository) importJobFromRecord(record sqlc.ImportJob) domain.ImportJob {
	item := domain.ImportJob{
		ID:           record.ID,
		Filename:     pgTextValue(record.Filename, ""),
		InputType:    record.InputType,
		Status:       domain.ImportJobStatus(record.Status),
		Progress:     int(record.Progress),
		ErrorMessage: pgTextPtr(record.ErrorMessage),
		CreatedAt:    pgTimestamptzValue(record.CreatedAt),
		StartedAt:    pgTimestamptzPtr(record.StartedAt),
		CompletedAt:  pgTimestamptzPtr(record.CompletedAt),
	}
	if len(record.Result) > 0 {
		var result domain.ImportPreviewResponse
		if err := json.Unmarshal(record.Result, &result); err == nil {
			item.Result = &result
		}
	}
	return item
}
