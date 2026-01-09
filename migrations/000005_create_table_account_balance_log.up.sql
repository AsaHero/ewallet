CREATE TABLE IF NOT EXISTS account_balance_log (
  id bigserial NOT NULL,
  user_id uuid NOT NULL,
  account_id uuid NOT NULL,
  transaction_id uuid NOT NULL,
  delta bigint NOT NULL,
  balance_before bigint NOT NULL,
  balance_after  bigint NOT NULL,
  occurred_at timestamp with time zone NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  PRIMARY KEY (id),
  CONSTRAINT account_balance_log_account_fk FOREIGN KEY (account_id) REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT account_balance_log_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT account_balance_log_transaction_fk FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS account_balance_log_account_tx_uniq ON account_balance_log(account_id, transaction_id);
CREATE INDEX IF NOT EXISTS account_balance_log_user_idx ON account_balance_log(user_id);
CREATE INDEX IF NOT EXISTS account_balance_log_account_idx ON account_balance_log(account_id);
CREATE INDEX IF NOT EXISTS account_balance_log_transaction_idx ON account_balance_log(transaction_id);

