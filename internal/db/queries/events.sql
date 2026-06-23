-- name: CreateEvent :one
INSERT INTO events (group_id, created_by, title, description, event_date, modality, link)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
