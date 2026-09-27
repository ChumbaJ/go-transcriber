ALTER TABLE transcriptions
ADD CONSTRAINT transcriptions_job_chunk_unique
UNIQUE (job_id, chunk_order);
