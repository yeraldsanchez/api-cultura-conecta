-- name: GetFocusTypes :many
SELECT * FROM focus_types;

-- name: CreateFocusType :one
INSERT INTO focus_types (name)
VALUES ($1)
RETURNING *;

-- name: AssignFocusTypeToUser :exec
INSERT INTO users_focus_types (profile_id, focus_type_id)
VALUES ($1, $2);

-- name: GetUserFocusTypes :many
SELECT ft.* FROM focus_types ft
JOIN users_focus_types uft ON ft.id = uft.focus_type_id
WHERE uft.profile_id = $1;

-- name: RemoveFocusTypeFromUser :exec
DELETE FROM users_focus_types
WHERE profile_id = $1 AND focus_type_id = $2;