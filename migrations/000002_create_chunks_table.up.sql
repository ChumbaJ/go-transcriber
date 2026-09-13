CREATE TABLE IF NOT EXISTS chunks (
  chunk_order INT NOT NULL,
  job_id  BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  addr  TEXT
);
