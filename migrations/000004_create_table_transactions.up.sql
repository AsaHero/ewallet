CREATE TABLE IF NOT EXISTS transactions(
    id uuid,
    user_id uuid NOT NULL,
    account_id uuid NOT NULL,
    category_id integer,
    type varchar(255) NOT NULL,
    status varchar(32),
    amount bigint NOT NULL,
    currency_code char(3) NOT NULL,
    original_amount bigint,
    original_currency_code char(3),
    fx_rate numeric(18, 6) precision,
    row_text text,
    performed_at timestamp with time zone,
    rejected_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT transactions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT transactions_category_id_fkey FOREIGN KEY (category_id) REFERENCES category(id) ON DELETE SET NULL ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS transactions_user_id_idx ON transactions(user_id);

CREATE INDEX IF NOT EXISTS transactions_account_id_idx ON transactions(account_id);

CREATE INDEX IF NOT EXISTS transactions_category_id_idx ON transactions(category_id);

CREATE INDEX IF NOT EXISTS transactions_created_at_idx ON transactions(created_at);

CREATE INDEX IF NOT EXISTS transactions_performed_at_idx ON transactions(performed_at);

