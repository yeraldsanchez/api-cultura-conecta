CREATE TABLE focus_types (
    id   SERIAL       PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE users_focus_types (
    profile_id    INTEGER NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    focus_type_id INTEGER NOT NULL REFERENCES focus_types(id) ON DELETE CASCADE,
    PRIMARY KEY (profile_id, focus_type_id)
);

ALTER TABLE user_profiles
    DROP COLUMN focus_type;

CREATE TABLE groups_focus_types (
    group_id      INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    focus_type_id INTEGER NOT NULL REFERENCES focus_types(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, focus_type_id)
);

ALTER TABLE groups
    DROP COLUMN focus_type;

