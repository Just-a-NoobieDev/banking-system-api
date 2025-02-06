CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed');

ALTER TABLE transactions ADD COLUMN status transaction_status;