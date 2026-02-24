CREATE TABLE IF NOT EXISTS transfers (
    id BIGSERIAL PRIMARY KEY,
    from_account_id BIGINT NOT NULL REFERENCES accounts(id),
    to_account_id BIGINT NOT NULL REFERENCES accounts(id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transfers_from_account_id ON transfers(from_account_id);

CREATE INDEX IF NOT EXISTS idx_transfers_to_account_id ON transfers(to_account_id);
