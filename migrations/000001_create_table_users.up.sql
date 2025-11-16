CREATE TABLE IF NOT EXISTS users(
    id uuid,
    tg_user_id bigint NOT NULL,
    first_name varchar(255),
    last_name varchar(255),
    username varchar(255),
    language_code char(2),
    currency_code char(3),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_tg_user_id_uindex ON users(tg_user_id);

