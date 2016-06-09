-- USERS
-- Password: lalala
INSERT INTO users (id, username, first_name, last_name, email, password) VALUES
    (1, 'admin', 'Nils', 'hesse', 'nphesse@gmail.com', '$2a$10$smTM4E3w/2U7PGRzyvUgV.EqD713TL7ABhrc3N9KEdqATCvRSqtky');

-- DEVICES
INSERT INTO devices (id, user_id, platform, name) VALUES
    ('519334e5-39a1-4cb2-bd21-64e053778c4c', 1, 'android', 'My Android Device');
INSERT INTO device_options (device_id, push_notifications) VALUES
    ('519334e5-39a1-4cb2-bd21-64e053778c4c', false);

INSERT INTO devices (id, user_id, platform, name) VALUES
    ('5a96d8df-b50f-4f63-a725-2c70fa63ed74', 1, 'chrome', 'My Chrome Device');
INSERT INTO device_options (device_id, push_notifications) VALUES
    ('5a96d8df-b50f-4f63-a725-2c70fa63ed74', false);

