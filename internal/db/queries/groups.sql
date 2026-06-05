-- name: CreateGroup :one
INSERT INTO groups (work_id, created_by, name, description, depth_level)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetGroupByID :one
SELECT g.id, g.work_id, g.created_by, g.name, g.description, g.depth_level, g.created_at,
       cw.title as work_title, up.name as created_by_name
FROM groups g
         JOIN cultural_works cw ON g.work_id = cw.id
         JOIN users u ON g.created_by = u.id
         LEFT JOIN user_profiles up ON u.id = up.user_id
WHERE g.id = $1;

-- name: AssignFocusTypeToGroup :exec
INSERT INTO groups_focus_types (group_id, focus_type_id)
VALUES ($1, $2);

-- name: GetGroupFocusTypes :many
SELECT ft.*
FROM focus_types ft
         JOIN groups_focus_types gft ON ft.id = gft.focus_type_id
WHERE gft.group_id = $1;