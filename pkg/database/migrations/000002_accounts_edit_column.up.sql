ALTER TABLE accounts
ADD CONSTRAINT accounts_title_unique UNIQUE (title);

ALTER TABLE accounts
ADD CONSTRAINT accounts_balance_positive CHECK (balance >= 0);

ALTER TABLE accounts
ALTER COLUMN balance TYPE BIGINT;