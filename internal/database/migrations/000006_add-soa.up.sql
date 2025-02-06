ALTER TABLE transactions ADD COLUMN user_id INT;
ALTER TABLE transactions ADD CONSTRAINT fk_transactions_users FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE statements DROP CONSTRAINT fk_statements_accounts;
ALTER TABLE statements DROP COLUMN account_id;
ALTER TABLE statements ADD COLUMN user_id INT;
ALTER TABLE statements ADD CONSTRAINT fk_statements_users FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE transactions DROP COLUMN description