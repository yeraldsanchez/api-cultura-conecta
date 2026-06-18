-- name: CreateGroup :one
INSERT INTO groups (work_id, created_by, name, description, depth_level)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetGroupByID :one
SELECT g.id,
       g.work_id,
       g.created_by,
       g.name,
       g.description,
       g.depth_level,
       g.created_at,
       cw.title as work_title,
       up.name  as created_by_name
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

-- name: CountGroups :one
SELECT COUNT(DISTINCT g.id)
FROM groups g
WHERE (sqlc.narg('work_id')::bigint IS NULL OR g.work_id = sqlc.narg('work_id'))
AND (sqlc.narg('name')::text IS NULL OR g.name ILIKE '%' || sqlc.narg('name') || '%')
AND (sqlc.narg('depth_level')::text IS NULL OR g.depth_level ILIKE '%' || sqlc.narg('depth_level') || '%')
AND (
    sqlc.narg('focus_type_ids')::integer[] IS NULL
    OR cardinality(sqlc.narg('focus_type_ids')::integer[]) = 0
    OR EXISTS (
        SELECT 1 FROM groups_focus_types gft
        WHERE gft.group_id = g.id
          AND gft.focus_type_id = ANY(sqlc.narg('focus_type_ids')::integer[])
    )
);

-- name: AddGroupMember :exec
INSERT INTO group_members (group_id, user_id, role)
VALUES ($1, $2, $3);

-- name: ListGroups :many
SELECT g.id,
       g.work_id,
       g.created_by,
       g.name,
       g.description,
       g.depth_level,
       g.created_at,
       cw.title as work_title,
       up.name  as created_by_name,
       JSONB_AGG(
           JSONB_BUILD_OBJECT('id', ft.id, 'name', ft.name)
       ) FILTER (WHERE ft.id IS NOT NULL) AS focus_types
FROM groups g
         JOIN cultural_works cw ON g.work_id = cw.id
         JOIN users u ON g.created_by = u.id
         LEFT JOIN user_profiles up ON u.id = up.user_id
         LEFT JOIN groups_focus_types gft ON gft.group_id = g.id
         LEFT JOIN focus_types ft ON ft.id = gft.focus_type_id
WHERE (sqlc.narg('work_id')::bigint IS NULL OR g.work_id = sqlc.narg('work_id'))
AND (sqlc.narg('name')::text IS NULL OR g.name ILIKE '%' || sqlc.narg('name') || '%')
AND (sqlc.narg('depth_level')::text IS NULL OR g.depth_level ILIKE '%' || sqlc.narg('depth_level') || '%')
AND (
    sqlc.narg('focus_type_ids')::integer[] IS NULL
    OR cardinality(sqlc.narg('focus_type_ids')::integer[]) = 0
    OR EXISTS (
        SELECT 1 FROM groups_focus_types gft2
        WHERE gft2.group_id = g.id
          AND gft2.focus_type_id = ANY(sqlc.narg('focus_type_ids')::integer[])
    )
)
GROUP BY g.id, cw.title, up.name
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: IsGroupMember :one
SELECT EXISTS (
    SELECT 1 FROM group_members
    WHERE group_id = $1 AND user_id = $2
);

-- name: CreatePost :one
INSERT INTO posts (group_id, user_id, content, has_spoiler, spoiler_progress)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
