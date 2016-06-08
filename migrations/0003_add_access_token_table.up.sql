CREATE TABLE access_tokens (
    id varchar(36) PRIMARY KEY,
    created_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id integer NOT NULL REFERENCES users (id) UNIQUE
);
