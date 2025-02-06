
ALTER TABLE transactions DROP CONSTRAINT fk_transactions_users;
ALTER TABLE transactions DROP COLUMN user_id;

ALTER TABLE statements DROP CONSTRAINT fk_statements_users;
ALTER TABLE statements DROP COLUMN user_id;
ALTER TABLE statements ADD COLUMN account_id INT;
ALTER TABLE statements ADD CONSTRAINT fk_statements_accounts FOREIGN KEY (account_id) REFERENCES accounts(id);

ALTER TABLE transactions ADD COLUMN description TEXT;