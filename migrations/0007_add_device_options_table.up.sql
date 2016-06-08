CREATE TABLE device_options (
    id SERIAL PRIMARY KEY,
    device_id varchar(36) NOT NULL REFERENCES devices (id),
    push_notifications boolean NOT NULL DEFAULT true
);
