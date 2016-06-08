CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    created_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id integer NOT NULL REFERENCES users (id),
    device_id varchar(36) NOT NULL REFERENCES devices (id),
    content_type integer NOT NULL CHECK (content_type > 0),
    title varchar(255),
    msg text,
    url text,
    file text
);
