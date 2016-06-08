CREATE TABLE devices (
    id varchar(36) PRIMARY KEY,
    created_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id integer NOT NULL REFERENCES users (id),
    platform varchar(16) NOT NULL,
    name varchar(32) NOT NULL
);
