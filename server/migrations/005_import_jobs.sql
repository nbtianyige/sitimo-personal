-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS import_jobs (
  id              text PRIMARY KEY,
  filename        text,
  input_type      text NOT NULL CHECK (input_type IN ('file_upload','latex_text')),
  status          text NOT NULL CHECK (status IN ('pending','processing','done','failed')),
  progress        int NOT NULL DEFAULT 0,
  result          jsonb,
  error_message   text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  started_at      timestamptz,
  completed_at    timestamptz
);

CREATE INDEX IF NOT EXISTS idx_import_jobs_status_created_at ON import_jobs(status, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_import_jobs_status_created_at;
DROP TABLE IF EXISTS import_jobs;
-- +goose StatementEnd
