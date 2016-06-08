CREATE TABLE friendships (
    id SERIAL PRIMARY KEY,
    created_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id integer NOT NULL REFERENCES users (id),
    has_friend_id integer NOT NULL REFERENCES users (id),
    UNIQUE(user_id, has_friend_id)
);
