-- +goose Up
CREATE TABLE workspace_lifecycle_steps (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id),
  step_name TEXT NOT NULL,
  desired_state TEXT NOT NULL,
  actual_state TEXT NOT NULL,
  provider_resource_id TEXT NOT NULL DEFAULT '',
  error_code TEXT NOT NULL DEFAULT '',
  last_checked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (workspace_id, step_name)
);

CREATE TABLE workspace_backups (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id),
  storage_id TEXT REFERENCES storage_volumes(id),
  provider_resource_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  created_by_user_id TEXT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE support_tickets
  ADD COLUMN priority TEXT NOT NULL DEFAULT 'normal',
  ADD COLUMN assignee_user_id TEXT REFERENCES users(id),
  ADD COLUMN failed_lifecycle_step TEXT NOT NULL DEFAULT '',
  ADD COLUMN fabric_error_code TEXT NOT NULL DEFAULT '',
  ADD COLUMN runtime_status JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN ledger_summary JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE approvals
  ADD COLUMN context JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE approvals DROP COLUMN context;

ALTER TABLE support_tickets
  DROP COLUMN ledger_summary,
  DROP COLUMN runtime_status,
  DROP COLUMN fabric_error_code,
  DROP COLUMN failed_lifecycle_step,
  DROP COLUMN assignee_user_id,
  DROP COLUMN priority;

DROP TABLE workspace_backups;
DROP TABLE workspace_lifecycle_steps;
