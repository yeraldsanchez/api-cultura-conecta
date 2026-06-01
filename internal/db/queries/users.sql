-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: CreateUserProfile :one
INSERT INTO user_profiles (user_id, depth_level)
VALUES ($1, $2)
RETURNING id;

-- name: GetUserProfileByUserId :one
SELECT u.id as user_id, upf.id as profile_id, u.email, upf.depth_level
FROM user_profiles upf
         JOIN users u ON upf.user_id = u.id
WHERE u.id = $1;