-- Drop indexes first
DROP INDEX IF EXISTS idx_transactions_account_id;
DROP INDEX IF EXISTS idx_transactions_created_at;
DROP INDEX IF EXISTS idx_accounts_user_id;
DROP INDEX IF EXISTS idx_statements_account_id;

-- Drop tables (in reverse order of creation to handle dependencies)
DROP TABLE IF EXISTS statements;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;

-- Drop custom types
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS currency_type;

