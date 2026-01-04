CREATE TABLE IF NOT EXISTS debts(
    id uuid,
    user_id uuid NOT NULL,
    transaction_id uuid NOT NULL,
    type varchar(255) NOT NULL CHECK (type IN ('borrow', 'lend')),
    status varchar(32) NOT NULL CHECK (status IN ('open', 'paid', 'cancelled')),
    amount bigint NOT NULL,
    currency_code char(3) NOT NULL,
    remind_at timestamp with time zone,
    paid_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone,
    PRIMARY KEY (id)
    CONSTRAINT debts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT debts_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE ON UPDATE CASCADE,
);

CREATE INDEX IF NOT EXISTS debts_user_id_type_status_idx ON debts(user_id, type, status);

CREATE INDEX IF NOT EXISTS debts_transaction_id_idx ON debts(transaction_id);

CREATE INDEX IF NOT EXISTS debts_remind_at_idx ON debts(remind_at);

