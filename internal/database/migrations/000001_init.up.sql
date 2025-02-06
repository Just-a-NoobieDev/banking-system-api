-- Create enum for currency types
CREATE TYPE currency_type AS ENUM ('USD', 'EUR', 'GBP');

-- Create enum for transaction types
CREATE TYPE transaction_type AS ENUM ('DEPOSIT', 'WITHDRAWAL', 'TRANSFER');

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    balance DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    currency currency_type NOT NULL DEFAULT 'USD',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    transaction_type transaction_type NOT NULL,
    description TEXT,
    reference_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS statements (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL,
    pdf_url TEXT NOT NULL,
    statement_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key constraints
ALTER TABLE transactions ADD CONSTRAINT fk_transactions_accounts FOREIGN KEY (account_id) REFERENCES accounts(id);
ALTER TABLE accounts ADD CONSTRAINT fk_accounts_users FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE statements ADD CONSTRAINT fk_statements_accounts FOREIGN KEY (account_id) REFERENCES accounts(id);

-- Add indexes for better performance
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_statements_account_id ON statements(account_id);




