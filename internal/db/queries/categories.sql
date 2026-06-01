-- name: GetCategories :many
SELECT * FROM categories;

-- name: CreateCategory :one
INSERT INTO categories (name)
VALUES ($1)
RETURNING *;

-- name: AssignInterestToUser :exec
INSERT INTO user_interests (profile_id, category_id)
VALUES ($1, $2);

-- name: GetUserInterests :many
SELECT c.* FROM categories c
JOIN user_interests ui ON c.id = ui.category_id
WHERE ui.profile_id = $1;