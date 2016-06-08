CREATE TABLE received_messages (
    id SERIAL PRIMARY KEY,
    created_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    device_id varchar(36) NOT NULL REFERENCES devices (id),
    message_id integer NOT NULL REFERENCES messages (id),
    unread boolean NOT NULL DEFAULT true
);
