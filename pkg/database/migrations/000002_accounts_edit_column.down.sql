ALTER TABLE accounts
ALTER COLUMN balance TYPE DECIMAL(10,2);

ALTER TABLE accounts
DROP CONSTRAINT IF EXISTS accounts_balance_positive;

ALTER TABLE accounts
DROP CONSTRAINT IF EXISTS accounts_title_unique;
