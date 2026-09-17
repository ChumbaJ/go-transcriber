CREATE TABLE IF NOT EXISTS jobs (
  id  BIGSERIAL PRIMARY KEY,
  status  TEXT NOT NULL DEFAULT 'pending'
          CHECK (status IN ('pending', 'completed', 'failed')),
  result_text TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
