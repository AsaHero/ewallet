CREATE TABLE IF NOT EXISTS accounts(
    id uuid,
    user_id uuid NOT NULL,
    name varchar(255) NOT NULL,
    balance bigint NOT NULL DEFAULT 0,
    is_default boolean NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone,
    PRIMARY KEY (id),
    CONSTRAINT accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS accounts_user_id_idx ON accounts(user_id);

