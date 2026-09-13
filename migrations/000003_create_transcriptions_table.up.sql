CREATE TABLE IF NOT EXISTS transcriptions (
  job_id BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  chunk_order INT NOT NULL,
  text TEXT NOT NULL,
  PRIMARY KEY(job_id, chunk_order)
);
