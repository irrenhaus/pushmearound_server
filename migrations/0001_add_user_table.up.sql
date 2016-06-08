CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_sign_in_at date NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username varchar(40) NOT NULL UNIQUE,
    first_name varchar(40),
    last_name varchar(40),
    email varchar(255) NOT NULL UNIQUE,
    email_confirmed boolean NOT NULL DEFAULT false,
    password varchar(255) NOT NULL
);
