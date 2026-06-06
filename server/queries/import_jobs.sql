-- name: GetImportJobByID :one
SELECT id, filename, input_type, status, progress, result, error_message,
  created_at, started_at, completed_at
FROM import_jobs
WHERE id = $1;

-- name: ListAllImportJobs :many
SELECT id, filename, input_type, status, progress, result, error_message,
  created_at, started_at, completed_at
FROM import_jobs
ORDER BY created_at DESC;

-- name: InsertImportJob :one
INSERT INTO import_jobs (
  id, filename, input_type, status, progress, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, filename, input_type, status, progress, result, error_message,
  created_at, started_at, completed_at;

-- name: UpdateImportJobStatus :exec
UPDATE import_jobs SET
  status = $2,
  progress = $3,
  started_at = $4,
  completed_at = $5,
  error_message = $6,
  result = $7
WHERE id = $1;

-- name: DeleteImportJob :exec
DELETE FROM import_jobs WHERE id = $1;
