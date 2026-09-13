CREATE TABLE IF NOT EXISTS jobs (
  id  BIGSERIAL PRIMARY KEY,
  status  TEXT NOT NULL DEFAULT 'pending'
          CHECK (status IN ('pending', 'completed', 'failed')),
  result_text TEXT DEFAULT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
