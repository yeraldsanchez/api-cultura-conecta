-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: CreateUserProfile :one
INSERT INTO user_profiles (user_id, name, depth_level)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserProfileByUserId :one
SELECT u.id as user_id, upf.id as profile_id, u.email, upf.name, upf.depth_level
FROM user_profiles upf
         JOIN users u ON upf.user_id = u.id
WHERE u.id = $1;

-- name: UpdateUserProfile :one
UPDATE user_profiles
SET
  name        = COALESCE($1::varchar, name),
  depth_level = COALESCE($2::varchar, depth_level)
WHERE user_id = $3
RETURNING id, user_id, name, depth_level, updated_at;