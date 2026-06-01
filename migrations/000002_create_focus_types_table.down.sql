ALTER TABLE groups
    ADD COLUMN focus_type VARCHAR(50) NOT NULL;

DROP TABLE IF EXISTS groups_focus_types;

ALTER TABLE user_profiles
    ADD COLUMN focus_type VARCHAR(50) NOT NULL;

DROP TABLE IF EXISTS users_focus_types;

DROP TABLE IF EXISTS focus_types;