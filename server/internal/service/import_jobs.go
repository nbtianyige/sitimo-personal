package service

import (
	"context"
	"encoding/json"

	"mathlib/server/internal/domain"
	"mathlib/server/internal/store"
)

func (s *Service) ListImportJobs(ctx context.Context, params ImportListParams) (domain.Paginated[domain.ImportJob], error) {
	return s.repo.ListImportJobs(ctx, store.ImportListOptions{
		Status:   params.Status,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
}

func (s *Service) GetImportJob(ctx context.Context, id string) (*domain.ImportJob, error) {
	return s.repo.GetImportJob(ctx, id)
}

func (s *Service) CreateImportJob(ctx context.Context, input ImportJobCreateInput) (*domain.ImportJob, error) {
	job, err := s.repo.CreateImportJobRecord(ctx, input.Filename, input.InputType)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateImportJobState(ctx, job.ID, domain.ImportJobStatusProcessing, 25, nil, nil)

	req := domain.ImportPreviewRequest{
		Files:    input.Files,
		Defaults: input.Defaults,
	}

	preview := s.PreviewBatchImport(req)
	if len(preview.Parsed) == 0 && len(preview.Errors) > 0 {
		errMsg := "导入解析失败，未识别到任何题目"
		resultBytes, _ := json.Marshal(preview)
		_ = s.repo.UpdateImportJobState(ctx, job.ID, domain.ImportJobStatusFailed, 100, resultBytes, &errMsg)
		return s.GetImportJob(ctx, job.ID)
	}

	_ = s.repo.UpdateImportJobState(ctx, job.ID, domain.ImportJobStatusProcessing, 60, nil, nil)

	_, commitErr := s.CommitBatchImport(ctx, preview.Parsed)
	resultBytes, _ := json.Marshal(preview)

	if commitErr != nil {
		errMsg := commitErr.Error()
		_ = s.repo.UpdateImportJobState(ctx, job.ID, domain.ImportJobStatusFailed, 100, resultBytes, &errMsg)
		return s.GetImportJob(ctx, job.ID)
	}

	_ = s.repo.UpdateImportJobState(ctx, job.ID, domain.ImportJobStatusDone, 100, resultBytes, nil)
	return s.GetImportJob(ctx, job.ID)
}

func (s *Service) DeleteImportJob(ctx context.Context, id string) error {
	return s.repo.DeleteImportJob(ctx, id)
}
