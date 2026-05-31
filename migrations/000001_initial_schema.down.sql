-- CulturaConecta — Rollback initial schema

-- Drop indexes
DROP INDEX IF EXISTS idx_groups_focus_depth;
DROP INDEX IF EXISTS idx_groups_work;
DROP INDEX IF EXISTS idx_events_date;
DROP INDEX IF EXISTS idx_events_group;
DROP INDEX IF EXISTS idx_posts_created_at;
DROP INDEX IF EXISTS idx_posts_group;
DROP INDEX IF EXISTS idx_group_members_user;
DROP INDEX IF EXISTS idx_cultural_works_category;
DROP INDEX IF EXISTS idx_user_interests_category;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS event_attendees;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS cultural_works;
DROP TABLE IF EXISTS user_interests;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS categories;
