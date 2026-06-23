-- name: CreateEvent :one
INSERT INTO events (group_id, created_by, title, description, event_date, modality, link)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEventsByGroup :many
SELECT * FROM events
WHERE group_id = $1
ORDER BY event_date ASC;

-- name: GetEventByID :one
SELECT * FROM events WHERE id = $1;

-- name: ConfirmAttendance :one
INSERT INTO event_attendees (event_id, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetEventAttendees :many
SELECT u.id          AS user_id,
       up.name       AS name,
       ea.confirmed_at
FROM event_attendees ea
         JOIN users u ON u.id = ea.user_id
         LEFT JOIN user_profiles up ON up.user_id = u.id
WHERE ea.event_id = $1
ORDER BY ea.confirmed_at;
