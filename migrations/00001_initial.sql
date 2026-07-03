-- +goose Up
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin')),
  status TEXT NOT NULL CHECK (status IN ('active', 'disabled')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organizations (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  billing_account_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE memberships (
  id TEXT PRIMARY KEY,
  organization_id TEXT NOT NULL REFERENCES organizations(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  role TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (organization_id, user_id)
);

CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  csrf_token_hash TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE billing_accounts (
  id TEXT PRIMARY KEY,
  owner_type TEXT NOT NULL CHECK (owner_type IN ('user', 'organization')),
  owner_id TEXT NOT NULL,
  balance_fen BIGINT NOT NULL DEFAULT 0 CONSTRAINT billing_accounts_balance_fen_check CHECK (balance_fen >= 0),
  frozen_fen BIGINT NOT NULL DEFAULT 0 CONSTRAINT billing_accounts_frozen_fen_check CHECK (frozen_fen >= 0),
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT billing_accounts_frozen_balance_check CHECK (frozen_fen <= balance_fen)
);

ALTER TABLE organizations
  ADD CONSTRAINT organizations_billing_account_id_fkey
  FOREIGN KEY (billing_account_id) REFERENCES billing_accounts(id);

CREATE TABLE wallet_transactions (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  amount_fen BIGINT NOT NULL CONSTRAINT wallet_transactions_amount_fen_check CHECK (amount_fen <> 0),
  kind TEXT NOT NULL,
  reason TEXT NOT NULL,
  actor_user_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wallet_holds (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  amount_fen BIGINT NOT NULL CONSTRAINT wallet_holds_amount_fen_check CHECK (amount_fen > 0),
  status TEXT NOT NULL CHECK (status IN ('active', 'released', 'debited')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE billing_ledger_entries (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  workspace_id TEXT,
  resource_type TEXT NOT NULL,
  resource_id TEXT,
  amount_fen BIGINT NOT NULL CONSTRAINT billing_ledger_entries_amount_fen_check CHECK (amount_fen <> 0),
  kind TEXT NOT NULL,
  description TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE manual_topups (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  amount_fen BIGINT NOT NULL CONSTRAINT manual_topups_amount_fen_check CHECK (amount_fen > 0),
  actor_user_id TEXT NOT NULL,
  note TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspace_packages (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  cpu INTEGER NOT NULL CONSTRAINT workspace_packages_cpu_check CHECK (cpu > 0),
  memory_gb INTEGER NOT NULL CONSTRAINT workspace_packages_memory_gb_check CHECK (memory_gb > 0),
  storage_gb INTEGER NOT NULL CONSTRAINT workspace_packages_storage_gb_check CHECK (storage_gb > 0),
  compute_hourly_fen BIGINT NOT NULL CONSTRAINT workspace_packages_compute_hourly_fen_check CHECK (compute_hourly_fen >= 0),
  storage_gb_month_fen BIGINT NOT NULL CONSTRAINT workspace_packages_storage_gb_month_fen_check CHECK (storage_gb_month_fen >= 0),
  available BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE compute_resources (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  package_id TEXT NOT NULL REFERENCES workspace_packages(id),
  provider_resource_id TEXT,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE storage_volumes (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  package_id TEXT NOT NULL REFERENCES workspace_packages(id),
  provider_resource_id TEXT,
  size_gb INTEGER NOT NULL CONSTRAINT storage_volumes_size_gb_check CHECK (size_gb > 0),
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE storage_attachments (
  id TEXT PRIMARY KEY,
  compute_id TEXT NOT NULL REFERENCES compute_resources(id),
  storage_id TEXT NOT NULL REFERENCES storage_volumes(id),
  mount_path TEXT NOT NULL DEFAULT '/data',
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspaces (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  name TEXT NOT NULL,
  package_id TEXT NOT NULL REFERENCES workspace_packages(id),
  compute_id TEXT REFERENCES compute_resources(id),
  storage_id TEXT REFERENCES storage_volumes(id),
  attachment_id TEXT REFERENCES storage_attachments(id),
  slug TEXT NOT NULL UNIQUE,
  state TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workspace_tokens (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id),
  token_hash TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('active', 'deleted')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX workspace_tokens_one_active_per_workspace_idx
  ON workspace_tokens(workspace_id)
  WHERE status = 'active';

CREATE INDEX workspace_tokens_token_hash_idx
  ON workspace_tokens(token_hash);

CREATE TABLE runtime_operations (
  id TEXT PRIMARY KEY,
  operation_type TEXT NOT NULL,
  actor_user_id TEXT NOT NULL,
  billing_account_id TEXT,
  workspace_id TEXT,
  stage TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
  error_code TEXT,
  error_message TEXT,
  request_id TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at TIMESTAMPTZ,
  CONSTRAINT runtime_operations_operation_type_request_id_key UNIQUE (operation_type, request_id)
);

CREATE TABLE support_tickets (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  workspace_id TEXT,
  subject TEXT NOT NULL,
  body TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'open',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notifications (
  id TEXT PRIMARY KEY,
  billing_account_id TEXT NOT NULL REFERENCES billing_accounts(id),
  user_id TEXT,
  level TEXT NOT NULL,
  message TEXT NOT NULL,
  read_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_events (
  id TEXT PRIMARY KEY,
  actor_user_id TEXT NOT NULL,
  action TEXT NOT NULL,
  object_type TEXT NOT NULL,
  object_id TEXT NOT NULL,
  request_id TEXT NOT NULL,
  result TEXT NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE receipts (
  id TEXT PRIMARY KEY,
  receipt_type TEXT NOT NULL,
  subject_type TEXT NOT NULL,
  subject_id TEXT NOT NULL,
  operation_id TEXT,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO workspace_packages (id, name, cpu, memory_gb, storage_gb, compute_hourly_fen, storage_gb_month_fen, available)
VALUES
  ('basic', 'Basic Workspace', 2, 4, 10, 39, 36, true),
  ('pro', 'Pro Workspace', 8, 16, 100, 309, 36, true);

-- +goose Down
DROP TABLE receipts;
DROP TABLE audit_events;
DROP TABLE notifications;
DROP TABLE support_tickets;
DROP TABLE runtime_operations;
DROP TABLE workspace_tokens;
DROP TABLE workspaces;
DROP TABLE storage_attachments;
DROP TABLE storage_volumes;
DROP TABLE compute_resources;
DROP TABLE workspace_packages;
DROP TABLE manual_topups;
DROP TABLE billing_ledger_entries;
DROP TABLE wallet_holds;
DROP TABLE wallet_transactions;
DROP TABLE sessions;
DROP TABLE memberships;
DROP TABLE organizations;
DROP TABLE billing_accounts;
DROP TABLE users;
