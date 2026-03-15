ALTER TABLE transfers 
ADD COLUMN idempotency_key VARCHAR(100) UNIQUE;