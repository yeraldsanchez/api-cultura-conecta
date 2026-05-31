-- CulturaConecta — Initial schema

-- ─────────────────────────────────────────
-- Cultural categories catalog
-- ─────────────────────────────────────────

CREATE TABLE categories (
    id   SERIAL       PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

-- ─────────────────────────────────────────
-- Users and profiles
-- ─────────────────────────────────────────

CREATE TABLE users (
    id            SERIAL       PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE user_profiles (
    id          SERIAL      PRIMARY KEY,
    user_id     INTEGER     NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    focus_type  VARCHAR(50) NOT NULL,
    depth_level VARCHAR(50) NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User interests: n-to-n relationship with categories
CREATE TABLE user_interests (
    profile_id  INTEGER NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (profile_id, category_id)
);

-- ─────────────────────────────────────────
-- Cultural works
-- ─────────────────────────────────────────

CREATE TABLE cultural_works (
    id          SERIAL       PRIMARY KEY,
    title       VARCHAR(255) NOT NULL UNIQUE,
    category_id INTEGER      NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    external_id VARCHAR(100),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────
-- Groups
-- ─────────────────────────────────────────

CREATE TABLE groups (
    id          SERIAL       PRIMARY KEY,
    work_id     INTEGER      NOT NULL REFERENCES cultural_works(id) ON DELETE RESTRICT,
    created_by  INTEGER      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    focus_type  VARCHAR(50)  NOT NULL,
    depth_level VARCHAR(50)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE group_members (
    group_id  INTEGER     NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id   INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role      VARCHAR(20) NOT NULL DEFAULT 'member'
                          CHECK (role IN ('member', 'moderator', 'admin')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

-- ─────────────────────────────────────────
-- Forum
-- ─────────────────────────────────────────

CREATE TABLE posts (
    id               SERIAL      PRIMARY KEY,
    group_id         INTEGER     NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id          INTEGER     NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    content          TEXT        NOT NULL,
    has_spoiler      BOOLEAN     NOT NULL DEFAULT FALSE,
    spoiler_progress VARCHAR(100),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────
-- Events
-- ─────────────────────────────────────────

CREATE TABLE events (
    id          SERIAL       PRIMARY KEY,
    group_id    INTEGER      NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_by  INTEGER      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    event_date  TIMESTAMPTZ  NOT NULL,
    modality    VARCHAR(20)  NOT NULL CHECK (modality IN ('in-person', 'virtual')),
    link        TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE event_attendees (
    event_id     INTEGER     NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id      INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    confirmed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, user_id)
);

-- ─────────────────────────────────────────
-- Indexes
-- ─────────────────────────────────────────

CREATE INDEX idx_user_interests_category ON user_interests(category_id);
CREATE INDEX idx_cultural_works_category ON cultural_works(category_id);
CREATE INDEX idx_group_members_user      ON group_members(user_id);
CREATE INDEX idx_posts_group             ON posts(group_id);
CREATE INDEX idx_posts_created_at        ON posts(created_at DESC);
CREATE INDEX idx_events_group            ON events(group_id);
CREATE INDEX idx_events_date             ON events(event_date);
CREATE INDEX idx_groups_work             ON groups(work_id);
CREATE INDEX idx_groups_focus_depth      ON groups(focus_type, depth_level);
